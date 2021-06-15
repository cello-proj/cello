package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	diff                  = "diff"
	defaultLocalSecureURI = "https://localhost:8443"
	sync                  = "sync"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client represents an API client.
type Client struct {
	authToken  string
	httpClient httpClient
	endpoint   string
}

// NewClient returns a new API client.
func NewClient(endpoint, authToken string) Client {
	// Automatically disable TLS verification if it's a local endpoint.
	// TODO handle this better.
	tr := &http.Transport{}
	if endpoint == defaultLocalSecureURI {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return Client{
		authToken:  authToken,
		endpoint:   endpoint,
		httpClient: &http.Client{Transport: tr},
	}
}

// GetWorkflowStatusOutput represents the output of GetWorkflowStatus.
type GetWorkflowStatusOutput struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Finished string `json:"finished"`
}

// GetWorkflowsOutput represents the output of GetWorkflowsOutput.
type GetWorkflowsOutput []string

// targetOperationInput represents the input to a targetOperation.
type targetOperationInput struct {
	Path        string
	ProjectName string
	SHA         string
	TargetName  string
	Type        string
}

// targetOperationOutput represents the output to a targetOperation.
type targetOperationOutput struct {
	WorkflowName string `json:"workflow_name"`
}

// targetOperationRequest represents a target operation request.
type targetOperationRequest struct {
	Path string `json:"path"`
	SHA  string `json:"sha"`
	Type string `json:"type"`
}

// DiffOutput represents the output for Diff.
type DiffOutput targetOperationOutput

// GetLogsOutput respresents the output for GetLogs.
type GetLogsOutput struct {
	Logs []string `json:"logs"`
}

// ExecuteWorkflowInput represents the input for ExecuteWorkflow.
type ExecuteWorkflowInput struct {
	Arguments            map[string][]string `json:"arguments"`
	EnvironmentVariables map[string]string   `json:"environment_variables"`
	Framework            string              `json:"framework"`
	Parameters           map[string]string   `json:"parameters"`
	ProjectName          string              `json:"project_name"`
	TargetName           string              `json:"target_name"`
	Type                 string              `json:"type"`
	WorkflowTemplateName string              `json:"workflow_template_name"`
}

// ExecuteWorkflowOutput represents the output for ExecuteWorkflow.
type ExecuteWorkflowOutput struct {
	WorkflowName string `json:"workflow_name"`
}

// SyncOutput represents the output for Sync.
type SyncOutput targetOperationOutput

// GetLogs gets the logs of a workflow.
func (c *Client) GetLogs(ctx context.Context, workflowName string) (GetLogsOutput, error) {
	url := fmt.Sprintf("%s/workflows/%s/logs", c.endpoint, workflowName)

	body, err := c.getRequest(ctx, url)
	if err != nil {
		return GetLogsOutput{}, err
	}

	var output GetLogsOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return GetLogsOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// StreamLogs streams the logs of a workflow.
// TODO how to handle the stream? Channel? maybe take a io.Writer/Closer?
func (c *Client) StreamLogs(ctx context.Context, workflowName string) (GetLogsOutput, error) {
	// TODO
	url := fmt.Sprintf("%s/workflows/%s/logstream", c.endpoint, workflowName)

	body, err := c.getRequest(ctx, url)
	if err != nil {
		return GetLogsOutput{}, err
	}

	var output GetLogsOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return GetLogsOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// GetWorkflowStatus gets the status of a workflow.
func (c *Client) GetWorkflowStatus(ctx context.Context, workflowName string) (GetWorkflowStatusOutput, error) {
	url := fmt.Sprintf("%s/workflows/%s", c.endpoint, workflowName)

	body, err := c.getRequest(ctx, url)
	if err != nil {
		return GetWorkflowStatusOutput{}, err
	}

	var output GetWorkflowStatusOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return GetWorkflowStatusOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// GetWorkflows gets the list of workflows for a project and target.
func (c *Client) GetWorkflows(ctx context.Context, project, target string) (GetWorkflowsOutput, error) {
	url := fmt.Sprintf("%s/projects/%s/targets/%s/workflows", c.endpoint, project, target)

	body, err := c.getRequest(ctx, url)
	if err != nil {
		return GetWorkflowsOutput{}, err
	}

	var output GetWorkflowsOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return GetWorkflowsOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// Diff submits a "diff" for the provided project target.
// TODO use 'input' struct?
func (c *Client) Diff(ctx context.Context, project, target, sha, path string) (DiffOutput, error) {
	targetOpInput := targetOperationInput{
		Path:        path,
		ProjectName: project,
		SHA:         sha,
		TargetName:  target,
		Type:        diff,
	}

	output, err := c.targetOperation(ctx, targetOpInput)
	if err != nil {
		return DiffOutput{}, err
	}

	return DiffOutput(output), nil
}

// ExecuteWorkflow submits a workflow execution request.
func (c *Client) ExecuteWorkflow(ctx context.Context, input ExecuteWorkflowInput) (ExecuteWorkflowOutput, error) {
	// TODO this should probably be refactored to be a different operation type
	// (like diff/sync).
	url := fmt.Sprintf("%s/workflows", c.endpoint)

	reqBody, err := json.Marshal(input)
	if err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("unable to create api request body, error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("unable to create api request: %w", err)
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return ExecuteWorkflowOutput{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var output ExecuteWorkflowOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// Sync submits a "sync" for the provided project target.
// TODO use 'input' struct?
func (c *Client) Sync(ctx context.Context, project, target, sha, path string) (SyncOutput, error) {
	targetOpInput := targetOperationInput{
		Path:        path,
		ProjectName: project,
		SHA:         sha,
		TargetName:  target,
		Type:        sync,
	}

	output, err := c.targetOperation(ctx, targetOpInput)
	if err != nil {
		return SyncOutput{}, err
	}

	return SyncOutput(output), nil
}

func (c *Client) getRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create api request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) targetOperation(ctx context.Context, input targetOperationInput) (targetOperationOutput, error) {
	url := fmt.Sprintf("%s/projects/%s/targets/%s/operations", c.endpoint, input.ProjectName, input.TargetName)

	targetReq := targetOperationRequest{
		Path: input.Path,
		SHA:  input.SHA,
		Type: input.Type,
	}

	reqBody, err := json.Marshal(targetReq)
	if err != nil {
		return targetOperationOutput{}, fmt.Errorf("unable to create api request body, error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return targetOperationOutput{}, fmt.Errorf("unable to create api request: %w", err)
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return targetOperationOutput{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return targetOperationOutput{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return targetOperationOutput{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var output targetOperationOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return targetOperationOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}
