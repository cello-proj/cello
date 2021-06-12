package api

import (
	"context"
	"fmt"
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
			wantErr:  fmt.Errorf("unable to create api request: parse \"\\f/workflows/foo\": net/url: invalid control character in URL"),
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

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			status, err := client.GetWorkflowStatus(context.Background(), "foo")

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, status, tt.want)
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
