package validations

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/distribution/distribution/reference"
	"github.com/go-playground/validator"
)

func NewValidator() (*validator.Validate, error) {
	validate := validator.New()

	//add custom validations
	customValidations := map[string]func(fl validator.FieldLevel) bool{
		"alphanumunderscore":            validateIsAlphaNumericUnderscore,
		"is_arn":                        validARN,
		"valid_target_type":             validTargetType,
		"valid_execute_container_image": validExecuteContainerImage,
		"valid_precontainer_image":      validPreContainerImage,
		"valid_argument":                validArgument,
	}

	for jsonTag, fnName := range customValidations {
		err := validate.RegisterValidation(jsonTag, fnName)
		if err != nil {
			return nil, err
		}
	}

	return validate, nil
}

// Initialize validator and validate struct
func ValidateStruct(s interface{}) error {
	validate, err := NewValidator()
	if err != nil {
		return err
	}
	if err := validate.Struct(s); err != nil {
		return structValidationErrors(err)
	}
	return nil
}

// Initialize validator and validate struct
func ValidateVar(name string, s interface{}, validation string) error {
	validate, err := NewValidator()
	if err != nil {
		return err
	}
	if err := validate.Var(s, validation); err != nil {
		return varValidationErrors(name, err)
	}
	return nil
}

// Vault does not allow for dashes
var isStringAlphaNumericUnderscore = regexp.MustCompile(`^([a-zA-Z])[a-zA-Z0-9_]*$`).MatchString

// ValidateValuer implements validator.CustomTypeFunc
func validateIsAlphaNumericUnderscore(fl validator.FieldLevel) bool {
	return isStringAlphaNumericUnderscore(fl.Field().String())
}

func validExecuteContainerImage(fl validator.FieldLevel) bool {
	image := fl.Field().MapIndex(reflect.ValueOf("execute_container_image_uri"))
	if image.IsValid() {
		return isValidImageURI(image.String())
	}
	// execute_container_image_uri key missing
	return false
}

func validPreContainerImage(fl validator.FieldLevel) bool {
	image := fl.Field().MapIndex(reflect.ValueOf("pre_container_image_uri"))
	if image.IsValid() {
		return isValidImageURI(image.String())
	}

	// pre_container_image_uri is not required
	return true
}

func validTargetType(fl validator.FieldLevel) bool {
	return fl.Field().String() == "aws_account"
}

// TODO long term, we should evaluate if hard coding in code is the right approach to
// specifying different argument types vs allowing dynmaic specification and
// interpolation in service/config.yaml
func validArgument(fl validator.FieldLevel) bool {
	for _, key := range fl.Field().MapKeys() {
		switch key.String() {
		case "execute", "init":
			return true
		}
	}
	return false
}

func validARN(fl validator.FieldLevel) bool {
	return arn.IsARN(fl.Field().String())
}

// Returns true, if the image uri is a valid container image uri
func isValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
}

func structValidationErrors(err error) error {
	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); ok {
		for _, validationError := range validationErrors {
			return fmt.Errorf("failed validation check for '%s' '%v'", validationError.Tag(), validationError.Field())
		}
	}
	return err
}

func varValidationErrors(name string, err error) error {
	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); ok {
		for _, validationError := range validationErrors {
			return fmt.Errorf("failed validation check for '%s' '%v'", name, validationError.Param())
		}
	}
	return err
}
