package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/argoproj-labs/argo-cloudops/internal/requests"
	"github.com/argoproj-labs/argo-cloudops/internal/responses"
	"github.com/argoproj-labs/argo-cloudops/service/internal/credentials"
	"github.com/argoproj-labs/argo-cloudops/service/internal/db"
	"github.com/argoproj-labs/argo-cloudops/service/internal/env"
	"github.com/argoproj-labs/argo-cloudops/service/internal/git"
	"github.com/argoproj-labs/argo-cloudops/service/internal/workflow"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

const (
	// #nosec
	testPassword = "D34DB33FD34DB33FD34DB33FD34DB33F"
)

type mockDB struct{}

func newMockDB() db.Client {
	return mockDB{}
}

func (d mockDB) CreateProjectEntry(ctx context.Context, pe db.ProjectEntry) error {
	if pe.ProjectID == "somedberror" {
		return fmt.Errorf("some db error")
	}

	return nil
}

func (d mockDB) ReadProjectEntry(ctx context.Context, project string) (db.ProjectEntry, error) {
	return db.ProjectEntry{}, nil
}

func (d mockDB) DeleteProjectEntry(ctx context.Context, project string) error {
	if project == "somedeletedberror" {
		return fmt.Errorf("some db error")
	}

	return nil
}

type mockGitClient struct{}

func newMockGitClient() git.Client {
	return mockGitClient{}
}

func (g mockGitClient) GetManifestFile(repository, commitHash, path string) ([]byte, error) {
	return loadFileBytes("TestCreateWorkflow/can_create_workflow.json")
}

type mockWorkflowSvc struct{}

func (m mockWorkflowSvc) Status(ctx context.Context, workflowName string) (*workflow.Status, error) {
	if workflowName == "WORKFLOW_ALREADY_EXISTS" {
		return &workflow.Status{Status: "success"}, nil
	}
	return &workflow.Status{Status: "failed"}, fmt.Errorf("workflow " + workflowName + " does not exist!")
}

func (m mockWorkflowSvc) Logs(ctx context.Context, workflowName string) (*workflow.Logs, error) {
	if workflowName == "WORKFLOW_ALREADY_EXISTS" {
		return nil, nil
	}
	return nil, fmt.Errorf("workflow " + workflowName + " does not exist!")
}

func (m mockWorkflowSvc) LogStream(ctx context.Context, workflowName string, w http.ResponseWriter) error {
	return nil
}

func (m mockWorkflowSvc) List(ctx context.Context) ([]string, error) {
	return []string{"project1-target1-abcde", "project2-target2-12345"}, nil
}

func (m mockWorkflowSvc) Submit(ctx context.Context, from string, parameters map[string]string) (string, error) {
	return "success", nil
}

func newMockProvider(a credentials.Authorization, env env.Vars, h http.Header, f credentials.VaultConfigFn, fn credentials.VaultSvcFn) (credentials.Provider, error) {
	return &mockCredentialsProvider{}, nil
}

type mockCredentialsProvider struct{}

func (m mockCredentialsProvider) GetToken() (string, error) {
	return testPassword, nil
}

func (m mockCredentialsProvider) CreateProject(name string) (string, string, error) {
	return "", "", nil
}

func (m mockCredentialsProvider) DeleteProject(name string) error {
	if name == "undeletableproject" {
		return fmt.Errorf("Some error occured deleting this project")
	}
	return nil
}

func (m mockCredentialsProvider) GetProject(proj string) (responses.GetProject, error) {
	if proj == "projectdoesnotexist" {
		return responses.GetProject{}, credentials.ErrNotFound
	}
	return responses.GetProject{Name: "project1"}, nil
}

func (m mockCredentialsProvider) CreateTarget(name string, req requests.CreateTarget) error {
	return nil
}

func (m mockCredentialsProvider) GetTarget(string, string) (responses.TargetProperties, error) {
	return responses.TargetProperties{}, nil
}

func (m mockCredentialsProvider) DeleteTarget(string, t string) error {
	if t == "undeletabletarget" {
		return fmt.Errorf("Some error occured deleting this target")
	}
	return nil
}

func (m mockCredentialsProvider) ListTargets(name string) ([]string, error) {
	if name == "undeletableprojecttargets" {
		return []string{"target1", "target2", "undeletabletarget"}, nil
	}
	return []string{}, nil
}

func (m mockCredentialsProvider) ProjectExists(name string) (bool, error) {
	existingProjects := []string{
		"projectalreadyexists",
		"undeletableprojecttargets",
		"undeletableproject",
		"somedeletedberror",
	}
	for _, existingProjects := range existingProjects {
		if name == existingProjects {
			return true, nil
		}
	}
	return false, nil
}

func (m mockCredentialsProvider) TargetExists(projectName, targetName string) (bool, error) {
	if targetName == "TARGET_ALREADY_EXISTS" {
		return true, nil
	}
	return false, nil
}

type test struct {
	name     string
	req      interface{}
	want     int
	body     string
	respFile string
	asAdmin  bool
	url      string
	method   string
}

func TestCreateProject(t *testing.T) {
	tests := []test{
		{
			name:     "fails to create project when not admin",
			req:      loadCreateProjectRequest(t, "TestCreateProject/fails_to_create_project_when_not_admin.json"),
			want:     http.StatusUnauthorized,
			respFile: "TestCreateProject/fails_to_create_project_when_not_admin_response.json",
			url:      "/projects",
			method:   "POST",
		},
		{
			name:    "can create project",
			req:     loadCreateProjectRequest(t, "TestCreateProject/can_create_project.json"),
			want:    http.StatusOK,
			asAdmin: true,
			url:     "/projects",
			method:  "POST",
		},
		{
			name:    "git repo must be valid",
			req:     loadCreateProjectRequest(t, "TestCreateProject/git_repo_must_be_valid.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects",
			method:  "POST",
		},
		{
			name:    "project name must be alphanumeric",
			req:     loadCreateProjectRequest(t, "TestCreateProject/project_name_must_be_alphanumeric.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects",
			method:  "POST",
		},
		{
			name:    "project name cannot be less than 4 characters",
			req:     loadCreateProjectRequest(t, "TestCreateProject/project_name_cannot_be_less_than_4_characters.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects",
			method:  "POST",
		},
		{
			name:    "project name cannot be greater than 32 characters",
			req:     loadCreateProjectRequest(t, "TestCreateProject/project_name_cannot_be_greater_than_32_characters.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects",
			method:  "POST",
		},
		{
			name:    "project name cannot already exist",
			req:     loadCreateProjectRequest(t, "TestCreateProject/project_name_cannot_already_exist.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects",
			method:  "POST",
		},
		{
			name:    "project fails to create db entry",
			req:     loadCreateProjectRequest(t, "TestCreateProject/project_fails_to_create_dbentry.json"),
			want:    http.StatusInternalServerError,
			asAdmin: true,
			url:     "/projects",
			method:  "POST",
		},
	}
	runTests(t, tests)
}

func TestDeleteProject(t *testing.T) {
	tests := []test{
		{
			name:    "fails to delete project when not admin",
			want:    http.StatusUnauthorized,
			asAdmin: false,
			url:     "/projects/projectalreadyexists",
			method:  "DELETE",
		},
		{
			name:    "can delete project",
			want:    http.StatusOK,
			asAdmin: true,
			url:     "/projects/projectalreadyexists",
			method:  "DELETE",
		},
		{
			name:    "fails to delete project if any targets exist",
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/undeletableprojecttargets",
			method:  "DELETE",
		},
		{
			name:    "fails to delete project",
			want:    http.StatusInternalServerError,
			asAdmin: true,
			url:     "/projects/undeletableproject",
			method:  "DELETE",
		},
		{
			name:    "fails to delete project db entry",
			want:    http.StatusInternalServerError,
			asAdmin: true,
			url:     "/projects/somedeletedberror",
			method:  "DELETE",
		},
	}
	runTests(t, tests)
}

func TestGetProject(t *testing.T) {
	tests := []test{
		{
			name:    "project exists, successful get project",
			want:    http.StatusOK,
			asAdmin: true,
			method:  "GET",
			url:     "/projects/project1",
		},
		{
			name:    "project does not exist",
			want:    http.StatusNotFound,
			asAdmin: true,
			method:  "GET",
			url:     "/projects/projectdoesnotexist",
		},
	}
	runTests(t, tests)
}

func TestCreateTarget(t *testing.T) {
	tests := []test{
		{
			name:    "fails to create target when not admin",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/fails_to_create_target_when_not_admin.json"),
			want:    http.StatusUnauthorized,
			asAdmin: false,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "can create target",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/can_create_target.json"),
			want:    http.StatusOK,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "target name must be alphanumeric",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/target_name_must_be_alphanumeric.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "target name cannot be less than 4 characters",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/target_name_cannot_be_less_than_4_characters.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "target name cannot be greater than 32 characters",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/target_name_cannot_be_greater_than_32_characters.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "target name cannot already exist",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/target_name_cannot_already_exist.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "type must be aws_account",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/type_must_be_aws_account.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "policy arns list length cannot be greater than 5",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/policy_arns_list_length_cannot_be_greater_than_5.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "policy arns are valid AWS arns",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/policy_arns_are_valid_aws_arns.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "role arn is valid AWS arn",
			req:     loadCreateTargetRequest(t, "TestCreateTarget/role_arn_is_valid_aws_arn.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets",
			method:  "POST",
		},
		{
			name:    "project must exist",
			req:     loadCreateWorkflowRequest(t, "TestCreateTarget/project_must_exist.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			url:     "/projects/PROJECT_DOES_NOT_EXIST/targets",
			method:  "POST",
		},
	}
	runTests(t, tests)
}

func TestDeleteTarget(t *testing.T) {
	tests := []test{
		{
			name:    "fails to delete target when not admin",
			want:    http.StatusUnauthorized,
			asAdmin: false,
			url:     "/projects/projectalreadyexists/targets/target1",
			method:  "DELETE",
		},
		{
			name:    "can delete target",
			want:    http.StatusOK,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets/target1",
			method:  "DELETE",
		},
		{
			name:    "target fails to delete",
			want:    http.StatusInternalServerError,
			asAdmin: true,
			url:     "/projects/projectalreadyexists/targets/undeletabletarget",
			method:  "DELETE",
		},
	}
	runTests(t, tests)
}

func TestCreateWorkflow(t *testing.T) {
	tests := []test{
		{
			name:    "can create workflows",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/can_create_workflow.json"),
			want:    http.StatusOK,
			body:    "{\"workflow_name\":\"success\"}\n",
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "execute_container_image_uri must be present",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/execute_container_image_uri_must_be_present.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "execute_container_image_uri must be valid",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/execute_container_image_uri_must_be_valid.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "framework must be present",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/framework_must_be_present.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "framework must be valid",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/framework_must_be_valid.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "parameters must be present",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/parameters_must_be_present.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "project name must be present",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/project_name_must_be_present.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "target name must be present",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/target_name_must_be_present.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "type must be present",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/type_must_be_present.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "type must be valid",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/type_must_be_valid.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "project must exist",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/project_must_exist.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
		{
			name:    "target must exist",
			req:     loadCreateWorkflowRequest(t, "TestCreateWorkflow/target_must_exist.json"),
			want:    http.StatusBadRequest,
			asAdmin: true,
			method:  "POST",
			url:     "/workflows",
		},
	}
	runTests(t, tests)
}

func TestCreateWorkflowFromGit(t *testing.T) {
	tests := []test{
		{
			name: "can create workflows",
			req: requests.CreateGitWorkflow{
				CommitHash: "sha123",
				Path:       "path/to/manifest.yaml",
				Type:       "sync",
			},
			want:    http.StatusOK,
			body:    "{\"workflow_name\":\"success\"}\n",
			asAdmin: true,
			method:  "POST",
			url:     "/projects/project1/targets/target1/operations",
		},
	}
	runTests(t, tests)
}

func TestGetWorkflow(t *testing.T) {
	tests := []test{
		{
			name:    "workflow exists, successful get workflow",
			want:    http.StatusOK,
			asAdmin: true,
			method:  "GET",
			url:     "/workflows/WORKFLOW_ALREADY_EXISTS",
		},
		{
			name:    "workflow does not exist",
			want:    http.StatusInternalServerError,
			asAdmin: true,
			method:  "GET",
			url:     "/workflows/WORKFLOW_DOES_NOT_EXIST",
		},
	}
	runTests(t, tests)
}

func TestGetWorkflowLogs(t *testing.T) {
	tests := []test{
		{
			name:    "successful get workflow logs",
			want:    http.StatusOK,
			asAdmin: true,
			method:  "GET",
			url:     "/workflows/WORKFLOW_ALREADY_EXISTS/logs",
		},
		{
			name:    "workflow does not exist",
			want:    http.StatusInternalServerError,
			asAdmin: true,
			method:  "GET",
			url:     "/workflows/WORKFLOW_DOES_NOT_EXIST/logs",
		},
	}
	runTests(t, tests)
}

func TestListWorkflows(t *testing.T) {
	tests := []test{
		{
			name:    "can get workflows",
			want:    http.StatusOK,
			asAdmin: true,
			method:  "GET",
			url:     "/projects/projects1/targets/target1/workflows",
		},
	}
	runTests(t, tests)
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name                  string
		endpoint              string // Used to cause a connection error.
		vaultStatusCode       int
		writeBadContentLength bool // Used to create response body error.
		wantResponseBody      string
		wantStatusCode        int
	}{
		{
			name:             "good_vault_200",
			vaultStatusCode:  http.StatusOK,
			wantResponseBody: "Health check succeeded\n",
			wantStatusCode:   http.StatusOK,
		},
		{
			name:             "good_vault_429",
			vaultStatusCode:  http.StatusTooManyRequests,
			wantResponseBody: "Health check succeeded\n",
			wantStatusCode:   http.StatusOK,
		},
		{
			// We want successful health check in this vault error scenario.
			name:                  "error_vault_read_response",
			vaultStatusCode:       http.StatusOK,
			writeBadContentLength: true,
			wantResponseBody:      "Health check succeeded\n",
			wantStatusCode:        http.StatusOK,
		},
		{
			name:             "error_vault_connection",
			endpoint:         string('\f'),
			wantResponseBody: "Health check failed\n",
			wantStatusCode:   http.StatusServiceUnavailable,
		},
		{
			name:             "error_vault_unhealthy_status_code",
			vaultStatusCode:  http.StatusInternalServerError,
			wantResponseBody: "Health check failed\n",
			wantStatusCode:   http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantURL := "/v1/sys/health"

			vaultSvc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != wantURL {
					http.NotFound(w, r)
				}

				if r.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}

				if tt.writeBadContentLength {
					w.Header().Set("Content-Length", "1")
				}

				w.WriteHeader(tt.vaultStatusCode)
			}))
			defer vaultSvc.Close()

			vaultEndpoint := vaultSvc.URL
			if tt.endpoint != "" {
				vaultEndpoint = tt.endpoint
			}

			h := handler{
				logger: log.NewNopLogger(),
				env: env.Vars{
					VaultAddress: vaultEndpoint,
				},
			}

			// Dummy request.
			req, err := http.NewRequest("", "", nil)
			if err != nil {
				assert.Nil(t, err)
			}

			resp := httptest.NewRecorder()

			h.healthCheck(resp, req)

			respResult := resp.Result()
			defer respResult.Body.Close()

			body, err := io.ReadAll(respResult.Body)
			assert.Nil(t, err)

			assert.Equal(t, tt.wantStatusCode, respResult.StatusCode)
			assert.Equal(t, tt.wantResponseBody, string(body))
		})
	}
}

// Serialize a type to JSON-encoded byte buffer.
func serialize(toMarshal interface{}) *bytes.Buffer {
	jsonStr, _ := json.Marshal(toMarshal)
	return bytes.NewBuffer(jsonStr)
}

// Run tests, checking the response status codes.
func runTests(t *testing.T, tests []test) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := executeRequest(tt.method, tt.url, serialize(tt.req), tt.asAdmin)
			if resp.StatusCode != tt.want {
				t.Errorf("Unexpected status code %d", resp.StatusCode)
			}

			if tt.body != "" {
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				defer resp.Body.Close()
				if err != nil {
					t.Errorf("Error loading body")
				}
				if tt.body != string(bodyBytes) {
					t.Errorf("Unexpected body '%s', expected '%s'", bodyBytes, tt.body)
				}
			}

			if tt.respFile != "" {
				wantBody, err := loadFileBytes(tt.respFile)
				if err != nil {
					t.Fatalf("unable to read response file '%s', err: '%s'", tt.respFile, err)
				}

				body, err := io.ReadAll(resp.Body)
				assert.Nil(t, err)

				defer resp.Body.Close()

				assert.JSONEq(t, string(wantBody), string(body))
			}
		})
	}
}

// Execute a generic HTTP request, making sure to add the appropriate authorization header.
func executeRequest(method string, url string, body *bytes.Buffer, asAdmin bool) *http.Response {
	config, err := loadConfig(testConfigPath)
	if err != nil {
		panic(fmt.Sprintf("Unable to load config %s", err))
	}

	h := handler{
		logger:                 log.NewNopLogger(),
		newCredentialsProvider: newMockProvider,
		argo:                   mockWorkflowSvc{},
		argoCtx: func(txID string) context.Context {
			return context.WithValue(context.Background(), requests.TxIDHeader, txID)
		},
		config:    config,
		gitClient: newMockGitClient(),
		env: env.Vars{
			AdminSecret: testPassword,
		},
		dbClient: newMockDB(),
	}

	var router = setupRouter(h)
	req, _ := http.NewRequest(method, url, body)
	authorizationHeader := "vault:user:" + testPassword
	if asAdmin {
		authorizationHeader = "vault:admin:" + testPassword
	}
	req.Header.Add("Authorization", authorizationHeader)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Result()
}

// Unmarshal a JSON file from the testdata directory into output.
func loadFileBytes(filename string) ([]byte, error) {
	file := filepath.Join("testdata", filename)
	fd, err := os.Open(file)
	if err != nil {
		return []byte{}, err
	}

	fileStat, err := fd.Stat()
	if err != nil {
		return []byte{}, err
	}

	fileContents := make([]byte, fileStat.Size())
	_, err = fd.Read(fileContents)

	return fileContents, err
}

// Unmarshal a JSON file from the testdata directory into output.
func loadJSON(t *testing.T, filename string, output interface{}) {
	file := filepath.Join("testdata", filename)
	fd, err := os.Open(file)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", file, err)
	}
	if err = json.NewDecoder(fd).Decode(&output); err != nil {
		t.Fatalf("failed to decode file %s: %v", file, err)
	}
}

// Load a createTargetRequest from the testdata directory.
func loadCreateTargetRequest(t *testing.T, filename string) (r *requests.CreateTarget) {
	loadJSON(t, filename, &r)
	return
}

// Load a createProjectRequest from the testdata directory.
func loadCreateProjectRequest(t *testing.T, filename string) (r *requests.CreateProject) {
	loadJSON(t, filename, &r)
	return
}

// Load a createWorkflowRequest from the testdata directory.
func loadCreateWorkflowRequest(t *testing.T, filename string) (r *requests.CreateWorkflow) {
	loadJSON(t, filename, &r)
	return
}
