package validations

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/distribution/distribution/reference"
	"github.com/go-playground/validator"
)

const (
	tagIsAlphaNumericUnderscore     = "is_alphanumunderscore"
	tagIsValidTargetType            = "is_valid_target_type"
	tagIsValidExecuteContainerImage = "is_valid_execute_container_image"
	tagIsValidPrecontainerImage     = "is_valid_precontainer_image"
	tagIsValidArgument              = "is_valid_argument"
	tagIsValidGitRepository         = "is_valid_git_repository"
	tagIsARN                        = "is_arn"
)

// NewValidator returns a validator.
func NewValidator() (*validator.Validate, error) {
	validate := validator.New()

	//add custom validations
	customValidations := map[string]func(fl validator.FieldLevel) bool{
		tagIsAlphaNumericUnderscore:     isAlphaNumericUnderscore,
		tagIsARN:                        isValidARN,
		tagIsValidTargetType:            isValidTargetType,
		tagIsValidExecuteContainerImage: isValidExecuteContainerImage,
		tagIsValidPrecontainerImage:     isValidPreContainerImage,
		tagIsValidArgument:              isValidArgument,
		tagIsValidGitRepository:         isValidGitRepository,
	}

	for jsonTag, fnName := range customValidations {
		if err := validate.RegisterValidation(jsonTag, fnName); err != nil {
			return nil, err
		}
	}

	return validate, nil
}

func Validate(validations ...func() error) error {
	for _, v := range validations {
		if err := v(); err != nil {
			return err
		}
	}

	return nil
}

func ValidateStruct2(input interface{}) error {
	customValidators := map[string]govalidator.CustomTypeValidator{
		"alphanumericunderscore2": isAlphaNumbericUnderscore2,
		"gitURI":                  isValidGitRepository2,
	}

	for k, v := range customValidators {
		if _, exists := govalidator.CustomTypeTagMap.Get(k); !exists {
			govalidator.CustomTypeTagMap.Set(k, v)
		}
	}

	_, err := govalidator.ValidateStruct(input)
	return err
}

// ValidateStruct initializes validators and validates struct.
func ValidateStruct(s interface{}) error {
	validate, err := NewValidator()
	if err != nil {
		return err
	}
	if err := validate.Struct(s); err != nil {
		return validationErrorMessage("structValidation", err)
	}
	return nil
}

// ValidateVar initializes validator and validates a var.
func ValidateVar(errorPrefix string, s interface{}, validation string) error {
	validate, err := NewValidator()
	if err != nil {
		return err
	}
	if err := validate.Var(s, validation); err != nil {
		return validationErrorMessage(errorPrefix, err)
	}
	return nil
}

func isAlphaNumericUnderscore(fl validator.FieldLevel) bool {
	// Vault does not allow dashes
	return regexp.MustCompile(`^([a-zA-Z])[a-zA-Z0-9_]*$`).MatchString(fl.Field().String())
}

// isAlphaNumbericUnderscore2
func isAlphaNumbericUnderscore2(field interface{}, kind interface{}) bool {
	// only handle strings
	switch s := field.(type) {
	case string:
		// Vault does not allow dashes and must start with alpha.
		pattern := `^([a-zA-Z])[a-zA-Z0-9_]*$`
		return regexp.MustCompile(pattern).MatchString(s)
	default:
		panic("unsupported field type for isAlphaNumbericUnderscore2")
	}
}

func isValidExecuteContainerImage(fl validator.FieldLevel) bool {
	image := fl.Field().MapIndex(reflect.ValueOf("execute_container_image_uri"))
	if image.IsValid() {
		return IsValidImageURI(image.String())
	}
	// execute_container_image_uri key missing
	return false
}

func isValidPreContainerImage(fl validator.FieldLevel) bool {
	image := fl.Field().MapIndex(reflect.ValueOf("pre_container_image_uri"))
	if image.IsValid() {
		return IsValidImageURI(image.String())
	}

	// pre_container_image_uri is not required
	return true
}

func isValidTargetType(fl validator.FieldLevel) bool {
	return fl.Field().String() == "aws_account"
}

// TODO long term, we should evaluate if hard coding in code is the right approach to
// specifying different argument types vs allowing dynamic specification and
// interpolation in service/config.yaml
func isValidArgument(fl validator.FieldLevel) bool {
	for _, key := range fl.Field().MapKeys() {
		switch key.String() {
		case "execute", "init":
			return true
		default:
			return false
		}
	}
	return true
}

// IsValidARN determines if the string is a valid AWS ARN.
func IsValidARN(s string) bool {
	return arn.IsARN(s)
}

func isValidARN(fl validator.FieldLevel) bool {
	return arn.IsARN(fl.Field().String())
}

// IsValidImageURI determines if the image URI is a valid container image URI
// format.
func IsValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
}

// isValidGitRepository2
func isValidGitRepository2(field interface{}, kind interface{}) bool {
	// only handle strings
	switch s := field.(type) {
	case string:
		pattern := `((git|ssh|https)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`
		return regexp.MustCompile(pattern).MatchString(s)
	default:
		panic("unsupported field type for isValidGitRepository")
	}
}

func isValidGitRepository(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`((git|ssh|https)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`).MatchString(fl.Field().String())
}

// Custom error messages
func validationErrorMessage(prefix string, err error) error {
	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); ok {
		validationError := validationErrors[0]
		switch validationError.Tag() {

		case tagIsARN:
			return fmt.Errorf("'%s' value '%v' is not a valid arn", validationError.Field(), validationError.Value())

		case tagIsValidTargetType:
			return fmt.Errorf("'%s' value '%v' is invalid, types supported:'aws_account'", validationError.Field(), validationError.Value())

		case tagIsAlphaNumericUnderscore:
			return fmt.Errorf("value '%v' is invalid, must only contain alpha numeric underscore characters", validationError.Value())

		case tagIsValidExecuteContainerImage:
			message := "an invalid container uri"
			if _, exist := validationError.Value().(map[string]string)["execute_container_image_uri"]; !exist {
				message = "required"
			}

			return fmt.Errorf("%s 'execute_container_image_uri' is %s", strings.ToLower(validationError.Field()), message)

		case tagIsValidPrecontainerImage:
			return fmt.Errorf("%s 'pre_container_image_uri' is an invalid container uri", strings.ToLower(validationError.Field()))

		case tagIsValidArgument:
			return fmt.Errorf("'%s' value '%v' is an invalid argument", validationError.Field(), validationError.Value())

		case tagIsValidGitRepository:
			return fmt.Errorf("%s is an invalid git uri", strings.ToLower(validationError.Field()))

		default:
			if validationError.Field() == "" {
				return fmt.Errorf("%s '%v'", prefix, validationError.Param())
			}
			return fmt.Errorf("%s is invalid", strings.ToLower(validationError.Field()))
		}

	}
	return err
}
