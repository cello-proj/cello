package requests

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/distribution/distribution/reference"
	"regexp"
)

// Convenience method for checking if a string is alphanumeric.
var isStringAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString

// Vault does not allow for dashes
var isStringAlphaNumericUnderscore = regexp.MustCompile(`^([a-zA-Z])[a-zA-Z0-9_]*$`).MatchString


func RunValidations(s interface{}, validations ...func(s interface{})error) error {
	for _, validation := range validations {
		if err := validation(s); err != nil {
			return err
		}
	}
	return nil
}

func ValidateWorkflowParameters(cwr interface{}) error {
	parameters := cwr.(CreateWorkflowRequest).Parameters
	if _, ok := parameters["execute_container_image_uri"]; !ok {
		return errors.New("parameters must include execute_container_image_uri")
	}

	if !isValidImageURI(parameters["execute_container_image_uri"]) {
		return errors.New("execute_container_image_uri must be a valid container uri")
	}

	if _, ok := parameters["pre_container_image_uri"]; ok {
		if !isValidImageURI(parameters["pre_container_image_uri"]) {
			return errors.New("pre_container_image_uri must be a valid container uri")
		}
	}

	return nil
}

// Validates a project name
func ValidateCreateWorkflowProjectName(cwr interface{}) error {
	projectName := cwr.(CreateWorkflowRequest).ProjectName
	return projectNameChecks(projectName)
}

// Validates a project name
func ValidateCreateProjectName(cwr interface{}) error {
	projectName := cwr.(CreateProjectRequest).Name
	return projectNameChecks(projectName)
}

func projectNameChecks(projectName string) error {
	if len(projectName) < 4 {
		return errors.New("project name must be greater than 3 characters")
	}

	if len(projectName) > 32 {
		return errors.New("project name must be less than 32 characters")
	}

	if !isStringAlphaNumeric(projectName) {
		return errors.New("project name must be alpha-numeric")
	}

	return nil
}

// Validates a target name
func ValidateCreateTargetName(cwr interface{}) error {
	targetName := cwr.(CreateTargetRequest).Name
	return targetNameChecks(targetName)
}
// Validates a target name
func ValidateCreateWorkflowTargetName(cwr interface{}) error {
	targetName := cwr.(CreateWorkflowRequest).TargetName
	return targetNameChecks(targetName)
}

func targetNameChecks(targetName string) error {
	if len(targetName) < 4 {
		return errors.New("target name must be greater than 3 characters")
	}

	if len(targetName) > 32 {
		return errors.New("target name must be less than 32 characters")
	}

	if !isStringAlphaNumericUnderscore(targetName) {
		return errors.New("target name must be alpha-numeric with underscores")
	}

	return nil
}

func ValidateCreateTargetProperties(ctr interface{}) error {
	properties := ctr.(CreateTargetRequest).Properties
	t := ctr.(CreateTargetRequest).Type
	if !isTypeAWSAccount(t) {
		return fmt.Errorf("type must be aws_account")
	}
	if len(properties.PolicyArns) > 5 {
		return fmt.Errorf("policy arns list length cannot be greater than 5")
	}
	for _, policyArn := range properties.PolicyArns {
		if !arn.IsARN(policyArn) {
			return fmt.Errorf("policy arn %s must be a valid arn", policyArn)
		}
	}
	if !arn.IsARN(properties.RoleArn) {
		return fmt.Errorf("role arn %s must be a valid arn", properties.RoleArn)
	}
	return nil
}


// Returns true, if the image uri is a valid container image uri
func isValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
}
func isTypeAWSAccount(t string) bool {
	return t == "aws_account"
}
