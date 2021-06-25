package requests

import "github.com/argoproj-labs/argo-cloudops/internal/validations"

// Create workflow request.
type CreateWorkflow struct {
	Arguments            map[string][]string `yaml:"arguments" json:"arguments"`
	EnvironmentVariables map[string]string   `yaml:"environment_variables" json:"environment_variables"`
	Framework            string              `yaml:"framework" json:"framework"`
	Parameters           map[string]string   `validate:"valid_execute_container_image,valid_precontainer_image" yaml:"parameters" json:"parameters"`
	ProjectName          string              `validate:"min=4,max=32,alphanum" yaml:"project_name" json:"project_name"`
	TargetName           string              `validate:"min=4,max=32,alphanumunderscore" yaml:"target_name" json:"target_name"`
	Type                 string              `yaml:"type" json:"type"`
	WorkflowTemplateName string              `yaml:"workflow_template_name" json:"workflow_template_name"`
}

func (req CreateWorkflow) Validate() error {
	return validations.ValidateStruct(req)
}

// Create workflow from git manifest request
type CreateGitWorkflow struct {
	Repository string `validate:"required" json:"repository"`
	CommitHash string `validate:"required,alphanum" json:"sha"`
	Path       string `validate:"required" json:"path"`
	Type       string `validate:"required" json:"type"`
}

func (req CreateGitWorkflow) Validate() error {
	return validations.ValidateStruct(req)
}

// Create target request.
type CreateTarget struct {
	Name       string           `validate:"min=4,max=32,alphanumunderscore" json:"name"`
	Properties TargetProperties `json:"properties"`
	Type       string           `validate:"valid_target_type" json:"type"`
}

func (req CreateTarget) Validate() error {
	return validations.ValidateStruct(req)
}

// Create project request.
type CreateProject struct {
	Name string `validate:"min=4,max=32,alphanum" json:"name"`
}

func (req CreateProject) Validate() error {
	return validations.ValidateStruct(req)
}

// Target properties for target requests.
type TargetProperties struct {
	CredentialType string   `json:"credential_type"`
	PolicyArns     []string `validate:"max=5,dive,is_arn" json:"policy_arns"`
	RoleArn        string   `validate:"is_arn" json:"role_arn"`
}
