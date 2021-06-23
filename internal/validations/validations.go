package validations

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/distribution/distribution/reference"
	"github.com/go-playground/validator"
)

func InitValidator() *validator.Validate {
	validate := validator.New()

	//add custom validations
	customValidations := map[string]func(fl validator.FieldLevel) bool{
		"alphanumunderscore":            ValidateIsAlphaNumbericUnderscore,
		"is_arn":                        ValidARN,
		"valid_target_type":             ValidTargetType,
		"valid_execute_container_image": ValidExecuteContainerImage,
		"valid_pre-container_image":     ValidPreContainerImage,
		"valid_argument":                ValidArgument,
		"valid_auth_header":             ValidateAuthHeader,
	}

	for jsonTag, fnName := range customValidations {
		validate.RegisterValidation(jsonTag, fnName)
	}

	return validate
}

// Vault does not allow for dashes
var isStringAlphaNumericUnderscore = regexp.MustCompile(`^([a-zA-Z])[a-zA-Z0-9_]*$`).MatchString

// ValidateValuer implements validator.CustomTypeFunc
func ValidateIsAlphaNumbericUnderscore(fl validator.FieldLevel) bool {
	return isStringAlphaNumericUnderscore(fl.Field().String())
}

func ValidExecuteContainerImage(fl validator.FieldLevel) bool {
	// found validates a key of "execute_container_image_uri" exists in map
	found := false
	for _, key := range fl.Field().MapKeys() {
		if key.String() == "execute_container_image_uri" {
			found = true
			if !isValidImageURI(fl.Field().MapIndex(key).String()) {
				return false // TODO should be false, true passes, other validations catching?
			}
		}
	}
	return found
}

func ValidPreContainerImage(fl validator.FieldLevel) bool {
	for _, key := range fl.Field().MapKeys() {
		if key.String() == "pre_container_image_uri" {
			if !isValidImageURI(fl.Field().MapIndex(key).String()) {
				return false
			}
		}
	}
	return true
}

func ValidTargetType(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case "aws_account":
		return true
	default:
		return false
	}
}

// TODO long term, we should evaluate if hard coding in code is the right approach to
// specifying different argument types vs allowing dynmaic specification and
// interpolation in service/config.yaml
func ValidArgument(fl validator.FieldLevel) bool {
	for _, key := range fl.Field().MapKeys() {
		switch key.String() {
		case "init":
			return true
		case "execute":
			return true
		default:
			return false
		}
	}
	return false
}

func ValidARN(fl validator.FieldLevel) bool {
	return arn.IsARN(fl.Field().String()) //fmt.Errorf("role arn %s must be a valid arn", roleArn)
}

// Returns true, if the image uri is a valid container image uri
func isValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
}

func ValidateAuthHeader(fl validator.FieldLevel) bool {
	auth := strings.SplitN(fl.Field().String(), ":", 3)
	for _, i := range auth {
		if i == "" {
			return false
		}
	}
	if len(auth) < 3 {
		return false
	}
	return true
}

func StructValidationErrors(err error) error {
	var errList []string
	for _, err := range err.(validator.ValidationErrors) {
		errList = append(errList, fmt.Sprintf("failed validation for check '%s' on '%v', value '%v' is not valid", err.Tag(), err.Field(), err.Value()))
	}
	return fmt.Errorf(strings.Join(errList, "\n"))
}

func VarValidationErrors(name string, err error) error {
	var errList []string
	for _, err := range err.(validator.ValidationErrors) {
		errList = append(errList, fmt.Sprintf(" %s failed validation, validation of '%s %s', for value %v", name, err.Tag(), err.Param(), err.Value()))
	}
	return fmt.Errorf(strings.Join(errList, "\n"))
}
