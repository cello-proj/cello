package requests

import (
	"errors"
	"fmt"
	"strings"

	"github.com/argoproj-labs/argo-cloudops/internal/validations"
)

// CreateWorkflow request.
// TODO: diff and sync should have separate validations/structs for validations
type CreateWorkflow struct {
	Arguments            map[string][]string `json:"arguments" yaml:"arguments"`
	EnvironmentVariables map[string]string   `json:"environment_variables" yaml:"environment_variables"`
	// We don't validate the specific framework as it's dynamic and can only be
	// done server side.
	Framework   string            `json:"framework" yaml:"framework" valid:"required~framework is required"`
	Parameters  map[string]string `json:"parameters" yaml:"parameters"`
	ProjectName string            `json:"project_name" yaml:"project_name" valid:"required~project_name is required,alphanum~project_name must be alphanumeric,stringlength(4|32)~project_name must be between 4 and 32 characters"`
	TargetName  string            `json:"target_name" yaml:"target_name" valid:"required~target_name is required,alphanumunderscore~target_name must be alphanumeric underscore,stringlength(4|32)~target_name must be between 4 and 32 characters"`
	// We don't validate the specific type as it's dynamic and can only be done
	// server side.
	Type                 string `json:"type" yaml:"type" valid:"required~type is required"`
	WorkflowTemplateName string `json:"workflow_template_name" yaml:"workflow_template_name" valid:"required~workflow_template_name is required"`
}

// Validate validates CreateWorkflow.
func (req CreateWorkflow) Validate(optionalValidations ...func() error) error {
	v := []func() error{
		func() error { return validations.ValidateStruct(req) },
		req.validateArguments,
		req.validateParameters,
	}
	v = append(v, optionalValidations...)

	return validations.Validate(v...)
}

// ValidateType is an optional validation should be passed as parameter to Validate().
func (req CreateWorkflow) ValidateType(types []string) func() error {
	return func() error {
		for _, t := range types {
			if req.Type == t {
				return nil
			}
		}

		return fmt.Errorf("type must be one of '%s'", strings.Join(types, " "))
	}
}

// validateParameters validates the Parameters.
// 'execute_container_image_uri' is required and the URI format will be
// validated.
// 'pre_container_image_uri' is optional. If it's provided, the URI format will
// be validated.
func (req CreateWorkflow) validateParameters() error {
	val, ok := req.Parameters["execute_container_image_uri"]
	if !ok {
		return errors.New("parameter execute_container_image_uri is required")
	}

	if !validations.IsValidImageURI(val) {
		return errors.New("parameter execute_container_image_uri must be a valid container uri")
	}

	if !validations.IsApprovedImageURI(val) {
		return errors.New("parameter execute_container_image_uri must be an approved image uri")
	}

	if val, ok := req.Parameters["pre_container_image_uri"]; ok {
		if !validations.IsValidImageURI(val) {
			return errors.New("parameter pre_container_image_uri must be a valid container uri")
		}

		if !validations.IsApprovedImageURI(val) {
			return errors.New("parameter pre_container_image_uri must be an approved image uri")
		}
	}

	return nil
}

// validateArguments validates the Arguments.
// If any Arguments are provided, they must be one of 'execute' or 'init'.
// TODO long term, we should evaluate if hard coding in code is the right
// approach to specifying different argument types vs allowing dynamic
// specification and interpolation in service/config.yaml
func (req CreateWorkflow) validateArguments() error {
	for k := range req.Arguments {
		if k != "execute" && k != "init" {
			return fmt.Errorf("arguments must be one of 'execute init'")
		}
	}

	return nil
}

// CreateGitWorkflow from git manifest request
type CreateGitWorkflow struct {
	CommitHash string `json:"sha" valid:"required~sha is required,alphanum~sha must be alphanumeric"`
	Path       string `json:"path" valid:"required~path is required"`
}

// Validate validates CreateGitWorkflow.
func (req CreateGitWorkflow) Validate() error {
	return validations.ValidateStruct(req)
}

// CreateTarget request.
type CreateTarget struct {
	Name       string           `json:"name" valid:"required~name is required,alphanumunderscore~name must be alphanumeric underscore,stringlength(4|32)~name must be between 4 and 32 characters"`
	Properties TargetProperties `json:"properties"`
	Type       string           `json:"type" valid:"required~type is required"`
}

// Validate validates CreateTarget.
func (req CreateTarget) Validate() error {
	v := []func() error{
		func() error { return validations.ValidateStruct(req) },
		func() error {
			if req.Type != "aws_account" {
				return errors.New("type must be one of 'aws_account'")
			}
			return nil
		},
		req.validateTargetProperties,
	}

	return validations.Validate(v...)
}

func (req CreateTarget) validateTargetProperties() error {
	if err := validations.ValidateStruct(req.Properties); err != nil {
		return err
	}

	if req.Properties.CredentialType != "assumed_role" {
		return errors.New("credential_type must be one of 'assumed_role'")
	}

	if !validations.IsValidARN(req.Properties.RoleArn) {
		return errors.New("role_arn must be a valid arn")
	}

	if len(req.Properties.PolicyArns) > 5 {
		return errors.New("policy_arns cannot be more than 5")
	}

	for _, arn := range req.Properties.PolicyArns {
		if !validations.IsValidARN(arn) {
			return errors.New("policy_arns contains an invalid arn")
		}
	}

	return nil
}

// CreateProject request.
type CreateProject struct {
	Name       string `json:"name" valid:"required~name is required,alphanum~name must be alphanumeric,stringlength(4|32)~name must be between 4 and 32 characters"`
	Repository string `json:"repository" valid:"required~repository is required"`
}

// Validate validates CreateProject.
func (req CreateProject) Validate() error {
	v := []func() error{
		func() error { return validations.ValidateStruct(req) },
		func() error {
			if !validations.IsValidGitURI(req.Repository) {
				return errors.New("repository must be a git uri")
			}
			return nil
		},
	}

	return validations.Validate(v...)
}

// TargetProperties for target requests.
type TargetProperties struct {
	CredentialType string   `json:"credential_type" valid:"required~credential_type is required"`
	PolicyArns     []string `json:"policy_arns"`
	PolicyDocument string   `json:"policy_document"`
	RoleArn        string   `json:"role_arn" valid:"required~role_arn is required"`
}

// TargetOperation represents a target operation request.
// TODO evaluate this vs. CreateGitWorkflow.
type TargetOperation struct {
	Path string `json:"path" valid:"required~path is required"`
	SHA  string `json:"sha" valid:"required~sha is required,alphanum~sha must be alphanumeric"`
	// We don't validate the specific type as it's dynamic and can only be done
	// server side.
	Type string `json:"type" valid:"required~type is required"`
}

// Validate validates TargetOperation.
func (req TargetOperation) Validate() error {
	return validations.ValidateStruct(req)
}
