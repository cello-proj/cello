package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cello-proj/cello/internal/requests"
	"github.com/cello-proj/cello/internal/responses"
)

const (
	diff                  = "diff"
	defaultLocalSecureURI = "https://localhost:8443"
	exec                  = "exec"
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
		// #nosec
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

// TargetOperationInput represents the input to a targetOperation.
type TargetOperationInput struct {
	Path        string
	ProjectName string
	SHA         string
	TargetName  string
}

// GetLogs gets the logs of a workflow.
func (c *Client) GetLogs(ctx context.Context, workflowName string) (responses.GetLogs, error) {
	url := fmt.Sprintf("%s/workflows/%s/logs", c.endpoint, workflowName)

	body, err := c.getRequest(ctx, url)
	if err != nil {
		return responses.GetLogs{}, err
	}

	var output responses.GetLogs
	if err := json.Unmarshal(body, &output); err != nil {
		return responses.GetLogs{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// StreamLogs streams the logs of a workflow starting after loggedBytes.
func (c *Client) StreamLogs(ctx context.Context, w io.Writer, workflowName string) error {
	var loggingCursorByte int64
	// When we receive a stream error we continue and retry processing up to 5 times keeping track of the byte we were logging.
	for i := 0; i < 10; i++ {
		err := c.streamLogsToWriterAtCursor(ctx, w, workflowName, &loggingCursorByte)
		if err != nil {
			// if connection idle timeout occurs retry
			if strings.Contains(err.Error(), "stream error: stream ID 1; INTERNAL_ERROR") || strings.Contains(err.Error(), "stream error: stream ID 3; INTERNAL_ERROR") {
				time.Sleep(time.Second * 10)
				continue
			}
			return err
		}
		break
	}
	return nil
}

func (c *Client) streamLogsToWriterAtCursor(ctx context.Context, w io.Writer, workflowName string, loggingCursorByte *int64) error {
	url := fmt.Sprintf("%s/workflows/%s/logstream", c.endpoint, workflowName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("unable to create api request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to make api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received unexpected status code: %d", resp.StatusCode)
	}

	// discard reader bytes till cursor byte number
	if _, err := io.CopyN(ioutil.Discard, resp.Body, *loggingCursorByte); err != nil {
		return err
	}
	writtenBytes, err := io.Copy(w, resp.Body)
	if err != nil {
		// retry call if we receive the stream error, increase the logging cursor byte by the amount we logged.
		*loggingCursorByte += writtenBytes
		return fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}
	return nil
}

// GetWorkflowStatus gets the status of a workflow.
func (c *Client) GetWorkflowStatus(ctx context.Context, workflowName string) (responses.GetWorkflowStatus, error) {
	url := fmt.Sprintf("%s/workflows/%s", c.endpoint, workflowName)

	body, err := c.getRequest(ctx, url)
	if err != nil {
		return responses.GetWorkflowStatus{}, err
	}

	var output responses.GetWorkflowStatus
	if err := json.Unmarshal(body, &output); err != nil {
		return responses.GetWorkflowStatus{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// GetWorkflows gets the list of workflows for a project and target.
func (c *Client) GetWorkflows(ctx context.Context, project, target string) (responses.GetWorkflows, error) {
	url := fmt.Sprintf("%s/projects/%s/targets/%s/workflows", c.endpoint, project, target)

	body, err := c.getRequest(ctx, url)
	if err != nil {
		return responses.GetWorkflows{}, err
	}

	var output responses.GetWorkflows
	if err := json.Unmarshal(body, &output); err != nil {
		return responses.GetWorkflows{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// Diff submits a "diff" for the provided project target.
func (c *Client) Diff(ctx context.Context, input TargetOperationInput) (responses.Diff, error) {
	output, err := c.targetOperation(ctx, input, diff)
	if err != nil {
		return responses.Diff{}, err
	}

	return responses.Diff(output), nil
}

// Exec submits an "exec" for the provided project target.
func (c *Client) Exec(ctx context.Context, input TargetOperationInput) (responses.Exec, error) {
	output, err := c.targetOperation(ctx, input, exec)
	if err != nil {
		return responses.Exec{}, err
	}

	return responses.Exec(output), nil
}

// ExecuteWorkflow submits a workflow execution request.
func (c *Client) ExecuteWorkflow(ctx context.Context, input requests.CreateWorkflow) (responses.ExecuteWorkflow, error) {
	// TODO this should probably be refactored to be a different operation type
	// (like diff/sync).
	url := fmt.Sprintf("%s/workflows", c.endpoint)

	reqBody, err := json.Marshal(input)
	if err != nil {
		return responses.ExecuteWorkflow{}, fmt.Errorf("unable to create api request body, error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return responses.ExecuteWorkflow{}, fmt.Errorf("unable to create api request: %w", err)
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return responses.ExecuteWorkflow{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return responses.ExecuteWorkflow{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return responses.ExecuteWorkflow{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var output responses.ExecuteWorkflow
	if err := json.Unmarshal(body, &output); err != nil {
		return responses.ExecuteWorkflow{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}

// Sync submits a "sync" for the provided project target.
func (c *Client) Sync(ctx context.Context, input TargetOperationInput) (responses.Sync, error) {
	output, err := c.targetOperation(ctx, input, sync)
	if err != nil {
		return responses.Sync{}, err
	}

	return responses.Sync(output), nil
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

func (c *Client) targetOperation(ctx context.Context, input TargetOperationInput, operationType string) (responses.TargetOperation, error) {
	url := fmt.Sprintf("%s/projects/%s/targets/%s/operations", c.endpoint, input.ProjectName, input.TargetName)

	targetReq := requests.TargetOperation{
		Path: input.Path,
		SHA:  input.SHA,
		Type: operationType,
	}

	if err := targetReq.Validate(); err != nil {
		return responses.TargetOperation{}, err
	}

	reqBody, err := json.Marshal(targetReq)
	if err != nil {
		return responses.TargetOperation{}, fmt.Errorf("unable to create api request body, error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return responses.TargetOperation{}, fmt.Errorf("unable to create api request: %w", err)
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return responses.TargetOperation{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return responses.TargetOperation{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return responses.TargetOperation{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var output responses.TargetOperation
	if err := json.Unmarshal(body, &output); err != nil {
		return responses.TargetOperation{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}
