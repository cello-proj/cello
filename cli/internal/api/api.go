package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultLocalSecureURI = "https://localhost:8443"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client represents an API client.
type Client struct {
	httpClient httpClient
	endpoint   string
}

// NewClient returns a new API client.
func NewClient(endpoint string) Client {
	// Automatically disable TLS verification if it's a local endpoint.
	// TODO handle this better.
	tr := &http.Transport{}
	if endpoint == defaultLocalSecureURI {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return Client{
		endpoint:   endpoint,
		httpClient: &http.Client{Transport: tr},
	}
}

// GetWorkflowStatus gets the status of a workflow.
func (c *Client) GetWorkflowStatus(ctx context.Context, name string) (GetWorkflowStatusResponse, error) {
	url := fmt.Sprintf("%s/workflows/%s", c.endpoint, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return GetWorkflowStatusResponse{}, fmt.Errorf("unable to create api request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GetWorkflowStatusResponse{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetWorkflowStatusResponse{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		return GetWorkflowStatusResponse{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var wfResp GetWorkflowStatusResponse
	if err := json.Unmarshal(body, &wfResp); err != nil {
		return wfResp, fmt.Errorf("unable to parse response: %w", err)
	}

	return wfResp, nil
}

type GetWorkflowStatusResponse struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Finished string `json:"finished"`
}
