package requests

import (
	"fmt"
	"strings"

	"github.com/argoproj-labs/argo-cloudops/internal/validations"
)

// CreateWorkflow request.
// TODO: diff and sync should have separate validations/structs for validations
type CreateWorkflow struct {
	Arguments            map[string][]string `validate:"is_valid_argument" yaml:"arguments" json:"arguments"`
	EnvironmentVariables map[string]string   `yaml:"environment_variables" json:"environment_variables"`
	Framework            string              `yaml:"framework" json:"framework"`
	Parameters           map[string]string   `validate:"is_valid_execute_container_image,is_valid_precontainer_image" yaml:"parameters" json:"parameters"`
	ProjectName          string              `validate:"min=4,max=32,alphanum" yaml:"project_name" json:"project_name"`
	TargetName           string              `validate:"min=4,max=32,is_alphanumunderscore" yaml:"target_name" json:"target_name"`
	Type                 string              `yaml:"type" json:"type"`
	WorkflowTemplateName string              `yaml:"workflow_template_name" json:"workflow_template_name"`
}

// ValidateType is an optional validation should be passed as parameter to Validate().
func (req CreateWorkflow) ValidateType(types []string) func() error {
	return func() error {
		validItems := strings.Join(types, " ")
		return validations.ValidateVar(
			"'type' must be one of",
			req.Type,
			fmt.Sprintf("oneof=%s", validItems),
		)
	}
}

// Validate validates CreateWorkflow.
func (req CreateWorkflow) Validate(optionalValidations ...func() error) error {
	for _, validation := range optionalValidations {
		if err := validation(); err != nil {
			return err
		}
	}
	return validations.ValidateStruct(req)
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
	Name       string `validate:"min=4,max=32,alphanum" json:"name"`
	Repository string `validate:"required,is_valid_git_repository" json:"repository"`
}

// Validate validates CreateProject.
func (req CreateProject) Validate() error {
	return validations.ValidateStruct(req)
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
