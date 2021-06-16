package requests


// Create workflow request.
type CreateWorkflowRequest struct {
	Arguments            map[string][]string `yaml:"arguments" json:"arguments"`
	EnvironmentVariables map[string]string   `yaml:"environment_variables" json:"environment_variables"`
	Framework            string              `yaml:"framework" json:"framework"`
	Parameters           map[string]string   `yaml:"parameters" json:"parameters"`
	ProjectName          string              `yaml:"project_name" json:"project_name"`
	TargetName           string              `yaml:"target_name" json:"target_name"`
	Type                 string              `yaml:"type" json:"type"`
	WorkflowTemplateName string              `yaml:"workflow_template_name" json:"workflow_template_name"`
}

// Create workflow from git manifest request
type CreateGitWorkflowRequest struct {
	Repository string `json:"repository"`
	CommitHash string `json:"sha"`
	Path       string `json:"path"`
	Type       string `json:"type"`
}

type CreateTargetRequest struct {
	Name       string                       `json:"name"`
	Properties TargetProperties `json:"properties"`
	Type       string                       `json:"type"`
}

type CreateProjectRequest struct {
	Name string `json:"name"`
}

type TargetProperties struct {
	CredentialType string   `json:"credential_type"`
	PolicyArns     []string `json:"policy_arns"`
	RoleArn        string   `json:"role_arn"`
}

