package responses

// Diff represents the responses for Diff.
type Diff TargetOperation

// Exec represents the responses for Exec.
type Exec TargetOperation

// ExecuteWorkflow represents the responses for ExecuteWorkflow.
type ExecuteWorkflow struct {
	WorkflowName string `json:"workflow_name"`
}

// GetLogs represents the responses for GetLogs.
type GetLogs struct {
	Logs []string `json:"logs"`
}

// GetProject represents the responses for GetProject.
type GetProject struct {
	Name string `json:"name"`
}

// GetTarget represents the responses for GetTarget.
type GetTarget struct {
	Name string `json:"name"`
}

// GetWorkflows represents the responses for GetWorkflows.
type GetWorkflows []string

// GetWorkflowStatus represents the responses for GetWorkflowStatus.
type GetWorkflowStatus struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Finished string `json:"finished"`
}

// Sync represents the responses for Sync.
type Sync TargetOperation

// TargetOperation represents the output to a targetOperation.
type TargetOperation struct {
	WorkflowName string `json:"workflow_name"`
}

// TargetProperties represents the responses for Targetproperties.
type TargetProperties struct {
	CredentialType string   `json:"credential_type"`
	PolicyArns     []string `json:"policy_arns"`
	PolicyDocument string   `json:"policy_document"`
	RoleArn        string   `json:"role_arn"`
}
