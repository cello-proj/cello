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

// targetOperationRequest represents the request for a target operation.
type targetOperationRequest struct {
	Path string `json:"path"`
	SHA  string `json:"sha"`
	Type string `json:"type"`
}

// targetOperationOutput represents the output to a target operation.
type targetOperationOutput struct {
	WorkflowName string `json:"workflow_name"`
}

// DiffOutput represents the output for Diff.
type DiffOutput targetOperationOutput

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

// GetWorkflowStatus gets the status of a workflow.
func (c *Client) GetWorkflowStatus(ctx context.Context, name string) (GetWorkflowStatusOutput, error) {
	url := fmt.Sprintf("%s/workflows/%s", c.endpoint, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return GetWorkflowStatusOutput{}, fmt.Errorf("unable to create api request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GetWorkflowStatusOutput{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetWorkflowStatusOutput{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		return GetWorkflowStatusOutput{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return GetWorkflowsOutput{}, fmt.Errorf("unable to create api request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GetWorkflowsOutput{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetWorkflowsOutput{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		return GetWorkflowsOutput{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var output GetWorkflowsOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return GetWorkflowsOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// TODO use 'input' struct?
// Diff submits a "diff" for the provided project target.
func (c *Client) Diff(ctx context.Context, project, target, sha, path string) (DiffOutput, error) {
	url := fmt.Sprintf("%s/projects/%s/targets/%s/operations", c.endpoint, project, target)

	targetReq := targetOperationRequest{
		Path: path,
		SHA:  sha,
		Type: diff,
	}

	reqBody, err := json.Marshal(targetReq)
	if err != nil {
		return DiffOutput{}, fmt.Errorf("unable to create api request body, error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return DiffOutput{}, fmt.Errorf("unable to create api request: %w", err)
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return DiffOutput{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DiffOutput{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return DiffOutput{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var output DiffOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return DiffOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
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

// TODO use 'input' struct?
// Diff submits a "diff" for the provided project target.
func (c *Client) Sync(ctx context.Context, project, target, sha, path string) (SyncOutput, error) {
	url := fmt.Sprintf("%s/projects/%s/targets/%s/operations", c.endpoint, project, target)

	targetReq := targetOperationRequest{
		Path: path,
		SHA:  sha,
		Type: sync,
	}

	reqBody, err := json.Marshal(targetReq)
	if err != nil {
		return SyncOutput{}, fmt.Errorf("unable to create api request body, error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return SyncOutput{}, fmt.Errorf("unable to create api request: %w", err)
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return SyncOutput{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SyncOutput{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return SyncOutput{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var output SyncOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return SyncOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}
