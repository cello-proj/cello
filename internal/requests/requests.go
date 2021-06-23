package requests

// Create workflow request.
type CreateWorkflowRequest struct {
	Arguments            map[string][]string `yaml:"arguments" json:"arguments"`
	EnvironmentVariables map[string]string   `yaml:"environment_variables" json:"environment_variables"`
	Framework            string              `yaml:"framework" json:"framework"`
	Parameters           map[string]string   `validate:"valid_execute_container_image,valid_pre-container_image" yaml:"parameters" json:"parameters"`
	ProjectName          string              `validate:"min=4,max=32,alphanum" yaml:"project_name" json:"project_name"`
	TargetName           string              `validate:"min=4,max=32,alphanumunderscore" yaml:"target_name" json:"target_name"`
	Type                 string              `yaml:"type" json:"type"`
	WorkflowTemplateName string              `yaml:"workflow_template_name" json:"workflow_template_name"`
}

// Create workflow from git manifest request
type CreateGitWorkflowRequest struct {
	Repository string `validate:"required" json:"repository"`
	CommitHash string `validate:"required,alphanum" json:"sha"`
	Path       string `validate:"required" json:"path"`
	Type       string `validate:"required" json:"type"`
}

// Create target request.
type CreateTargetRequest struct {
	Name       string           `validate:"min=4,max=32,alphanumunderscore" json:"name"`
	Properties TargetProperties `json:"properties"`
	Type       string           `validate:"valid_target_type" json:"type"`
}

// Create project request.
type CreateProjectRequest struct {
	Name string `validate:"min=4,max=32,alphanum" json:"name"`
}

// Target properties for target requests.
type TargetProperties struct {
	CredentialType string   `json:"credential_type"`
	PolicyArns     []string `validate:"max=5,dive,is_arn" json:"policy_arns"`
	RoleArn        string   `validate:"is_arn" json:"role_arn"`
}
