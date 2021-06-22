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
	validate.RegisterValidation("alphanumunderscore", ValidateIsAlphaNumbericUnderscore)
	validate.RegisterValidation("is_arn", ValidARN)
	validate.RegisterValidation("valid_target_type", ValidTargetType)
	validate.RegisterValidation("valid_execute_container_image", ValidExecuteContainerImage)
	validate.RegisterValidation("valid_pre-container_image", ValidPreContainerImage)
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

func ValidARN(fl validator.FieldLevel) bool {
	return arn.IsARN(fl.Field().String()) //fmt.Errorf("role arn %s must be a valid arn", roleArn)
}

// Returns true, if the image uri is a valid container image uri
func isValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
}

func ValidateAuthHeader(authorizationHeader string) error {
	auth := strings.SplitN(authorizationHeader, ":", 3)
	for _, i := range auth {
		if i == "" {
			return fmt.Errorf("invalid authorization header provided")
		}
	}
	if len(auth) < 3 {
		return fmt.Errorf("invalid authorization header provided")
	}
	return nil
}

func ValidationErrors(err error) error {
	var errList []string
	for _, err := range err.(validator.ValidationErrors) {
		errList = append(errList, fmt.Sprintf("failed validation %s, validation error found in %v does not have valid value %v", err.Tag(), err.Field(), err.Value()))
	}
	return fmt.Errorf(strings.Join(errList, "\n"))
}
