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
	Framework            string              `json:"framework" yaml:"framework"`
	Parameters           map[string]string   `json:"parameters" yaml:"parameters"`
	// TODO do we need to validate this as we've already done so on project creation? won't we return a project not found if it's invalid?
	ProjectName string `json:"project_name" yaml:"project_name" valid:"alphanum~project_name must be alphanumeric,stringlength(4|32)~project_name must be between 4 and 32 characters"`
	// TODO do we need to validate this as we've already done so on project creation? won't we return a project not found if it's invalid?
	TargetName           string `json:"target_name" yaml:"target_name" valid:"alphanumericunderscore2~target_name must be alphanumericunderscore,stringlength(4|32)~target_name must be between 4 and 32 characters"`
	Type                 string `json:"type" yaml:"type"`
	WorkflowTemplateName string `json:"workflow_template_name" yaml:"workflow_template_name"`
}

// Validate validates CreateWorkflow.
func (req CreateWorkflow) Validate(optionalValidations ...func() error) error {
	v := []func() error{
		func() error { return validations.ValidateStruct2(req) },
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

	if val, ok := req.Parameters["pre_container_image_uri"]; ok {
		if !validations.IsValidImageURI(val) {
			return errors.New("parameter pre_container_image_uri must be a valid container uri")
		}
	}

	return nil
}

// validateArguments validates the Arguments.
// The valid Arguments cases are:
// * no arguments
// * both 'execute' and 'init'
func (req CreateWorkflow) validateArguments() error {
	if len(req.Arguments) == 0 {
		return nil
	}

	if len(req.Arguments) > 2 {
		return fmt.Errorf("arguments must be one of 'execute init'")
	}

	for k := range req.Arguments {
		if k != "execute" && k != "init" {
			return fmt.Errorf("arguments must be one of 'execute init'")
		}
	}

	return nil
}

// CreateGitWorkflow from git manifest request
type CreateGitWorkflow struct {
	CommitHash string `validate:"required,alphanum" json:"sha"`
	Path       string `validate:"required" json:"path"`
	Type       string `validate:"required" json:"type"`
}

// Validate validates CreateGitWorkflow.
func (req CreateGitWorkflow) Validate() error {
	return validations.ValidateStruct(req)
}

// CreateTarget request.
type CreateTarget struct {
	Name       string           `validate:"min=4,max=32,is_alphanumunderscore" json:"name"`
	Properties TargetProperties `json:"properties"`
	Type       string           `validate:"is_valid_target_type" json:"type"`
}

// Validate validates CreateTarget.
func (req CreateTarget) Validate() error {
	return validations.ValidateStruct(req)
}

// CreateProject request.
type CreateProject struct {
	Name       string `json:"name" valid:"alphanum~name must be alphanumeric,stringlength(4|32)~name must be between 4 and 32 characters"`
	Repository string `json:"repository" valid:"required,gitURI~repository must be a git uri" validate:"required,is_valid_git_repository"`
}

// Validate validates CreateProject.
func (req CreateProject) Validate() error {
	return validations.ValidateStruct2(req)
}

// TargetProperties for target requests.
type TargetProperties struct {
	CredentialType string   `json:"credential_type"`
	PolicyArns     []string `validate:"max=5,dive,is_arn" json:"policy_arns"`
	PolicyDocument string   `json:"policy_document"`
	RoleArn        string   `validate:"is_arn" json:"role_arn"`
}

// TargetOperation represents a target operation request.
type TargetOperation struct {
	Path string `validate:"required" json:"path"`
	SHA  string `validate:"required,alphanum" json:"sha"`
	Type string `validate:"required,oneof=diff sync" json:"type"`
}

// Validate validates TargetOperation.
func (req TargetOperation) Validate() error {
	return validations.ValidateStruct(req)
}
