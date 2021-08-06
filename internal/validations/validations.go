package validations

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

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

func isValidExecuteContainerImage(fl validator.FieldLevel) bool {
	image := fl.Field().MapIndex(reflect.ValueOf("execute_container_image_uri"))
	if image.IsValid() {
		return isValidImageURI(image.String())
	}
	// execute_container_image_uri key missing
	return false
}

func isValidPreContainerImage(fl validator.FieldLevel) bool {
	image := fl.Field().MapIndex(reflect.ValueOf("pre_container_image_uri"))
	if image.IsValid() {
		return isValidImageURI(image.String())
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

func isValidARN(fl validator.FieldLevel) bool {
	return arn.IsARN(fl.Field().String())
}

// Returns true, if the image uri is a valid container image uri
func isValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
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
			return fmt.Errorf("'%s' value '%v' is an invalid git repository name, repo name must be in the format of 'git@url.com:owner/repo.git'", validationError.Field(), validationError.Value())

		default:
			if validationError.Field() == "" {
				return fmt.Errorf("%s '%v'", prefix, validationError.Param())
			}
			return fmt.Errorf("'%s' '%v'", validationError.Field(), validationError.Field())
		}

	}
	return err
}
