package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWorkflowStatus(t *testing.T) {
	tests := []struct {
		name                  string
		apiRespBody           []byte
		apiRespStatusCode     int
		endpoint              string          // Used to create new request error.
		mockHTTPClient        *mockHTTPClient // Only used when needed.
		writeBadContentLength bool            // Used to create response body error.
		want                  GetWorkflowStatusResponse
		wantErr               error
	}{
		{
			name:              "good",
			apiRespBody:       readFile(t, "get_workflow_status_good.json"),
			apiRespStatusCode: http.StatusOK,
			want: GetWorkflowStatusResponse{
				Name:     "foo-name",
				Status:   "succeeded",
				Created:  "1257891000",
				Finished: "1257894000",
			},
		},
		{
			name:              "error non-200 response",
			apiRespBody:       []byte("boom"),
			apiRespStatusCode: http.StatusInternalServerError,
			wantErr:           fmt.Errorf("received unexpected status code: 500, body: boom"),
		},
		{
			name:              "error non-json response",
			apiRespBody:       []byte("boom"),
			apiRespStatusCode: 200,
			wantErr:           fmt.Errorf("unable to parse response: invalid character 'b' looking for beginning of value"),
		},
		{
			name:     "error creating http request",
			endpoint: string('\f'),
			wantErr:  fmt.Errorf("unable to create api request: parse \"\\f/workflows/project1\": net/url: invalid control character in URL"),
		},
		{
			name:           "error making http request",
			mockHTTPClient: &mockHTTPClient{errDo: fmt.Errorf("boom")},
			wantErr:        fmt.Errorf("unable to make api call: boom"),
		},
		{
			name:                  "error reading body",
			apiRespBody:           nil,
			apiRespStatusCode:     http.StatusOK,
			writeBadContentLength: true,
			wantErr:               fmt.Errorf("error reading response body. status code: %d, error: unexpected EOF", http.StatusOK),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantURL := "/workflows/project1"

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != wantURL {
					http.NotFound(w, r)
				}

				// TODO handle verb. Better way to wire up a handler?

				if tt.writeBadContentLength {
					w.Header().Set("Content-Length", "1")
				}
				w.WriteHeader(tt.apiRespStatusCode)
				fmt.Fprint(w, string(tt.apiRespBody))
			}))
			defer server.Close()

			client := Client{
				endpoint:   server.URL,
				httpClient: &http.Client{},
			}

			if tt.endpoint != "" {
				client.endpoint = tt.endpoint
			}

			if tt.mockHTTPClient != nil {
				client.httpClient = tt.mockHTTPClient
			}

			status, err := client.GetWorkflowStatus(context.Background(), "project1")

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, status, tt.want)
			}
		})
	}
}

func TestGetWorkflows(t *testing.T) {
	tests := []struct {
		name                  string
		apiRespBody           []byte
		apiRespStatusCode     int
		endpoint              string          // Used to create new request error.
		mockHTTPClient        *mockHTTPClient // Only used when needed.
		writeBadContentLength bool            // Used to create response body error.
		want                  GetWorkflowsResponse
		wantErr               error
	}{
		{
			name:              "good",
			apiRespBody:       readFile(t, "get_workflows_good.json"),
			apiRespStatusCode: http.StatusOK,
			want:              GetWorkflowsResponse{"foo", "bar", "baz"},
		},
		{
			name:              "error non-200 response",
			apiRespBody:       []byte("boom"),
			apiRespStatusCode: http.StatusInternalServerError,
			wantErr:           fmt.Errorf("received unexpected status code: 500, body: boom"),
		},
		{
			name:              "error non-json response",
			apiRespBody:       []byte("boom"),
			apiRespStatusCode: 200,
			wantErr:           fmt.Errorf("unable to parse response: invalid character 'b' looking for beginning of value"),
		},
		{
			name:     "error creating http request",
			endpoint: string('\f'),
			wantErr:  fmt.Errorf("unable to create api request: parse \"\\f/projects/project1/targets/target1/workflows\": net/url: invalid control character in URL"),
		},
		{
			name:           "error making http request",
			mockHTTPClient: &mockHTTPClient{errDo: fmt.Errorf("boom")},
			wantErr:        fmt.Errorf("unable to make api call: boom"),
		},
		{
			name:                  "error reading body",
			apiRespBody:           nil,
			apiRespStatusCode:     http.StatusOK,
			writeBadContentLength: true,
			wantErr:               fmt.Errorf("error reading response body. status code: %d, error: unexpected EOF", http.StatusOK),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantURL := "/projects/project1/targets/target1/workflows"

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != wantURL {
					http.NotFound(w, r)
				}

				// TODO handle verb. Better way to wire up a handler?

				if tt.writeBadContentLength {
					w.Header().Set("Content-Length", "1")
				}
				w.WriteHeader(tt.apiRespStatusCode)
				fmt.Fprint(w, string(tt.apiRespBody))
			}))
			defer server.Close()

			client := Client{
				endpoint:   server.URL,
				httpClient: &http.Client{},
			}

			if tt.endpoint != "" {
				client.endpoint = tt.endpoint
			}

			if tt.mockHTTPClient != nil {
				client.httpClient = tt.mockHTTPClient
			}

			// TODO update var
			status, err := client.GetWorkflows(context.Background(), "project1", "target1")

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, status, tt.want)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	tests := []struct {
		name                  string
		apiRespBody           []byte
		apiRespStatusCode     int
		endpoint              string          // Used to create new request error.
		mockHTTPClient        *mockHTTPClient // Only used when needed.
		writeBadContentLength bool            // Used to create response body error.
		want                  DiffResponse
		wantAPIReqBody        []byte // TODO use string?
		wantErr               error
	}{
		{
			name:              "good",
			apiRespBody:       readFile(t, "diff_response_good.json"),
			apiRespStatusCode: http.StatusOK,
			want: DiffResponse{
				WorkflowName: "workflow1",
			},
			wantAPIReqBody: readFile(t, "diff_request_good.json"),
		},
		{
			name:              "error non-200 response",
			apiRespBody:       []byte("boom"),
			apiRespStatusCode: http.StatusInternalServerError,
			wantAPIReqBody:    readFile(t, "diff_request_good.json"),
			wantErr:           fmt.Errorf("received unexpected status code: 500, body: boom"),
		},
		{
			name:              "error non-json response",
			apiRespBody:       []byte("boom"),
			apiRespStatusCode: 200,
			wantAPIReqBody:    readFile(t, "diff_request_good.json"),
			wantErr:           fmt.Errorf("unable to parse response: invalid character 'b' looking for beginning of value"),
		},
		{
			name:     "error creating http request",
			endpoint: string('\f'),
			wantErr:  fmt.Errorf("unable to create api request: parse \"\\f/projects/project1/targets/target1/operations\": net/url: invalid control character in URL"),
		},
		{
			name:           "error making http request",
			mockHTTPClient: &mockHTTPClient{errDo: fmt.Errorf("boom")},
			wantErr:        fmt.Errorf("unable to make api call: boom"),
		},
		{
			name:                  "error reading body",
			apiRespBody:           nil,
			apiRespStatusCode:     http.StatusOK,
			writeBadContentLength: true,
			wantAPIReqBody:        readFile(t, "diff_request_good.json"),
			wantErr:               fmt.Errorf("error reading response body. status code: %d, error: unexpected EOF", http.StatusOK),
		},
		// TODO any others? generate our request?
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authToken := "secret1234"
			wantURL := "/projects/project1/targets/target1/operations"

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != wantURL {
					http.NotFound(w, r)
				}

				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}

				if tt.writeBadContentLength {
					w.Header().Set("Content-Length", "1")
				}

				// Make sure the request we received is what we want
				body, err := io.ReadAll(r.Body)
				r.Body.Close()

				assert.Nil(t, err, "unable to read request body")

				assert.JSONEq(t, string(body), string(tt.wantAPIReqBody))
				assert.Equal(t, r.Header.Get("Authorization"), authToken)

				w.WriteHeader(tt.apiRespStatusCode)
				fmt.Fprint(w, string(tt.apiRespBody))
			}))
			defer server.Close()

			client := Client{
				authToken:  authToken,
				endpoint:   server.URL,
				httpClient: &http.Client{},
			}

			if tt.endpoint != "" {
				client.endpoint = tt.endpoint
			}

			if tt.mockHTTPClient != nil {
				client.httpClient = tt.mockHTTPClient
			}

			diff, err := client.Diff(
				context.Background(),
				"project1",
				"target1",
				"7fa96067f580a20c3908f5b872377181091ffaec",
				"./prod/target1.yaml",
			)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, diff, tt.want)
			}
		})
	}
}

type mockHTTPClient struct {
	errDo error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{}, m.errDo
}

func readFile(t *testing.T, fileName string) []byte {
	data, err := os.ReadFile(filepath.Join("testdata", fileName))
	if err != nil {
		t.Fatalf("unable to read test file '%s', error: %s", fileName, err)
	}
	return data
}
