package responses

// CreateToken represents the output for CreateToken.
type CreateToken struct {
	CreatedAt string `json:"created_at"`
	Token     string `json:"token"`
	TokenID   string `json:"token_id"`
}

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
	Name       string `json:"name"`
	Repository string `json:"repository"`
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

// ListTokens represents the responses for ListTokens.
type ListTokens struct {
	CreatedAt string `json:"created_at"`
	ProjectID string `json:"project,omitempty"`
	TokenID   string `json:"token_id"`
}

// Sync represents the responses for Sync.
type Sync TargetOperation

// TargetOperation represents the output to a targetOperation.
type TargetOperation struct {
	WorkflowName string `json:"workflow_name"`
}
