package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cello-proj/cello/internal/responses"
	"github.com/cello-proj/cello/internal/types"
	"github.com/cello-proj/cello/service/internal/credentials"
	"github.com/cello-proj/cello/service/internal/db"
	"github.com/cello-proj/cello/service/internal/env"
	"github.com/cello-proj/cello/service/internal/git"
	"github.com/cello-proj/cello/service/internal/workflow"
	th "github.com/cello-proj/cello/service/test/testhelpers"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	upper "github.com/upper/db/v4"
)

const (
	// #nosec
	testPassword        = "D34DB33FD34DB33FD34DB33FD34DB33F"
	userAuthHeader      = "vault:user:" + testPassword
	invalidAuthHeader   = "bad auth header"
	adminAuthHeader     = "vault:admin:" + testPassword
	projectDoesNotExist = "projectdoesnotexist"
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

func (d mockDB) CreateTokenEntry(ctx context.Context, project string, secretAccessor string) (db.TokenEntry, error) {
	if project == "tokendberror" {
		return db.TokenEntry{}, fmt.Errorf("token db error")
	}

	token := db.TokenEntry{
		CreatedAt: "2022-06-21T14:56:10.341066-07:00",
		ProjectID: project,
		TokenID:   secretAccessor,
	}
	return token, nil
}

func (d mockDB) ListTokenEntries(ctx context.Context, project string) ([]db.TokenEntry, error) {
	if project == projectDoesNotExist {
		return []db.TokenEntry{}, upper.ErrNoMoreRows
	}

	if project == "projectreaderror" {
		return []db.TokenEntry{}, errors.New("error reading DB")
	}

	if project == "projectlisttokenserror" {
		return []db.TokenEntry{}, errors.New("error reading DB")
	}

	return []db.TokenEntry{}, upper.ErrNoMoreRows
}

func (d mockDB) ReadProjectEntry(ctx context.Context, project string) (db.ProjectEntry, error) {
	if project == projectDoesNotExist {
		return db.ProjectEntry{}, upper.ErrNoMoreRows
	}

	if project == "projectreaderror" {
		return db.ProjectEntry{}, errors.New("error reading DB")
	}

	return db.ProjectEntry{}, nil
}

func (d mockDB) DeleteProjectEntry(ctx context.Context, project string) error {
	if project == "somedeletedberror" {
		return fmt.Errorf("some db error")
	}

	return nil
}

func (d mockDB) DeleteTokenEntry(ctx context.Context, token string) error {
	return nil
}

func (d mockDB) ReadTokenEntry(ctx context.Context, token string) (db.TokenEntry, error) {
	return db.TokenEntry{}, nil
}

type mockGitClient struct{}

func newMockGitClient() git.Client {
	return mockGitClient{}
}

func (g mockGitClient) GetManifestFile(repository, commitHash, path string) ([]byte, error) {
	return loadFileBytes("TestCreateWorkflow/can_create_workflow_request.json")
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

func (m mockWorkflowSvc) Submit(ctx context.Context, from string, parameters map[string]string, labels map[string]string) (string, error) {
	return "wf-123456", nil
}

func newMockProvider(a credentials.Authorization, env env.Vars, h http.Header, f credentials.VaultConfigFn, fn credentials.VaultSvcFn) (credentials.Provider, error) {
	return &mockCredentialsProvider{}, nil
}

type mockCredentialsProvider struct{}

func (m mockCredentialsProvider) DeleteProjectToken(projectName, tokenID string) error {
	return nil
}

func (m mockCredentialsProvider) GetProjectToken(projectName, tokenID string) (types.ProjectToken, error) {
	return types.ProjectToken{}, nil
}

func (m mockCredentialsProvider) GetToken() (string, error) {
	return testPassword, nil
}

func (m mockCredentialsProvider) CreateProject(name string) (string, string, string, error) {
	return "role-id", "secret", "secret-id-accessor", nil
}

func (m mockCredentialsProvider) CreateToken(name string) (string, string, string, error) {
	return "role-id", "secret", "secret-id-accessor", nil
}

func (m mockCredentialsProvider) DeleteProject(name string) error {
	if name == "undeletableproject" {
		return fmt.Errorf("Some error occured deleting this project")
	}
	return nil
}

func (m mockCredentialsProvider) GetProject(proj string) (responses.GetProject, error) {
	if proj == projectDoesNotExist {
		return responses.GetProject{}, credentials.ErrNotFound
	}
	return responses.GetProject{Name: "project1"}, nil
}

func (m mockCredentialsProvider) CreateTarget(name string, req types.Target) error {
	return nil
}

func (m mockCredentialsProvider) GetTarget(project string, target string) (types.Target, error) {
	if target == "targetdoesnotexist" {
		return types.Target{}, credentials.ErrNotFound
	}
	return types.Target{
		Name: "TARGET",
		Type: "aws_account",
		Properties: types.TargetProperties{
			CredentialType: "assumed_role",
			PolicyArns: []string{
				"arn:aws:iam::012345678901:policy/test-policy",
			},
			PolicyDocument: "{ \"Version\": \"2012-10-17\", \"Statement\": [ { \"Effect\": \"Allow\", \"Action\": \"s3:ListBuckets\", \"Resource\": \"*\" } ] }",
			RoleArn:        "arn:aws:iam::012345678901:role/test-role",
		},
	}, nil
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
		"tokendberror",
		"projectnotokens",
		"projectreaderror",
		"projectlisttokenserror",
	}
	for _, existingProjects := range existingProjects {
		if name == existingProjects {
			return true, nil
		}
	}
	return false, nil
}

func (m mockCredentialsProvider) TargetExists(projectName, targetName string) (bool, error) {
	if targetName == "TARGET_EXISTS" {
		return true, nil
	}
	return false, nil
}

func (m mockCredentialsProvider) UpdateTarget(projectName string, target types.Target) error {
	return nil
}

type test struct {
	name       string
	req        interface{}
	want       int
	body       string
	respFile   string
	authHeader string
	url        string
	method     string
	dbMock     *th.DBClientMock
	cpMock     *th.CredsProviderMock
}

func TestCreateProject(t *testing.T) {
	tests := []test{
		{
			name:       "fails to create project when not admin",
			req:        loadJSON(t, "TestCreateProject/fails_to_create_project_when_not_admin_request.json"),
			want:       http.StatusUnauthorized,
			authHeader: userAuthHeader,
			respFile:   "TestCreateProject/fails_to_create_project_when_not_admin_response.json",
			url:        "/projects",
			method:     "POST",
		},
		{
			name:       "can create project",
			req:        loadJSON(t, "TestCreateProject/can_create_project_request.json"),
			want:       http.StatusOK,
			respFile:   "TestCreateProject/can_create_project_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects",
			method:     "POST",
		},
		{
			name:       "bad request",
			req:        loadJSON(t, "TestCreateProject/bad_request.json"),
			want:       http.StatusBadRequest,
			respFile:   "TestCreateProject/bad_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects",
			method:     "POST",
		},
		{
			name:       "project name cannot already exist",
			req:        loadJSON(t, "TestCreateProject/project_name_cannot_already_exist.json"),
			want:       http.StatusBadRequest,
			authHeader: adminAuthHeader,
			url:        "/projects",
			method:     "POST",
		},
		{
			name:       "project fails to create db entry",
			req:        loadJSON(t, "TestCreateProject/project_fails_to_create_dbentry.json"),
			want:       http.StatusInternalServerError,
			authHeader: adminAuthHeader,
			url:        "/projects",
			method:     "POST",
		},
	}
	runTests(t, tests)
}

func TestCreateToken(t *testing.T) {
	tests := []test{

		{
			name:       "can create token",
			req:        loadJSON(t, "TestCreateToken/request.json"),
			want:       http.StatusOK,
			respFile:   "TestCreateToken/can_create_token_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/undeletableprojecttargets/tokens",
			method:     "POST",
		},
		{
			name:       "project does not exist",
			req:        loadJSON(t, "TestCreateToken/request.json"),
			want:       http.StatusNotFound,
			respFile:   "TestCreateToken/project_does_not_exist.json",
			authHeader: adminAuthHeader,
			url:        "/projects/project1234/tokens",
			method:     "POST",
		},
		{
			name:       "fails to create token when not admin",
			req:        loadJSON(t, "TestCreateToken/request.json"),
			want:       http.StatusUnauthorized,
			respFile:   "TestCreateToken/fails_to_create_token_when_not_admin_response.json",
			authHeader: userAuthHeader,
			url:        "/projects/undeletableprojecttargets/tokens",
			method:     "POST",
		},
		{
			name:       "token fails to create db entry",
			req:        loadJSON(t, "TestCreateToken/request.json"),
			want:       http.StatusInternalServerError,
			authHeader: adminAuthHeader,
			url:        "/projects/tokendberror/tokens",
			method:     "POST",
		},
	}
	runTests(t, tests)
}

func TestGetTarget(t *testing.T) {
	tests := []test{
		{
			name:       "fails to get target when not admin",
			want:       http.StatusUnauthorized,
			respFile:   "TestGetTarget/fails_to_get_target_when_not_admin_response.json",
			authHeader: userAuthHeader,
			url:        "/projects/undeletableprojecttargets/targets/target1",
			method:     "GET",
		},
		{
			name:       "can get target",
			want:       http.StatusOK,
			respFile:   "TestGetTarget/can_get_target_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/undeletableprojecttargets/targets/TARGET_EXISTS",
			method:     "GET",
		},
		{
			name:       "target does not exist",
			want:       http.StatusNotFound,
			respFile:   "TestGetTarget/target_does_not_exist_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/undeletableprojecttargets/targets/targetdoesnotexist",
			method:     "GET",
		},
	}
	runTests(t, tests)
}

func TestListTargets(t *testing.T) {
	tests := []test{
		{
			name:       "fails to list targets when not admin",
			want:       http.StatusUnauthorized,
			respFile:   "TestListTargets/fails_to_list_targets_when_not_admin_response.json",
			authHeader: userAuthHeader,
			url:        "/projects/undeletableprojecttargets/targets",
			method:     "GET",
		},
		{
			name:       "can list targets",
			want:       http.StatusOK,
			respFile:   "TestListTargets/can_get_target_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/undeletableprojecttargets/targets",
			method:     "GET",
		},
		{
			name:       "project not found",
			want:       http.StatusNotFound,
			respFile:   "TestListTargets/project_not_found_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/badproject/targets",
			method:     "GET",
		},
		{
			name:       "no targets",
			want:       http.StatusOK,
			respFile:   "TestListTargets/no_targets_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets",
			method:     "GET",
		},
	}
	runTests(t, tests)
}

func TestDeleteProject(t *testing.T) {
	tests := []test{
		{
			name:       "fails to delete project when not admin",
			want:       http.StatusUnauthorized,
			authHeader: userAuthHeader,
			url:        "/projects/projectalreadyexists",
			method:     "DELETE",
		},
		{
			name:       "can delete project",
			want:       http.StatusOK,
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists",
			method:     "DELETE",
		},
		{
			name:       "fails to delete project if any targets exist",
			want:       http.StatusBadRequest,
			authHeader: adminAuthHeader,
			url:        "/projects/undeletableprojecttargets",
			method:     "DELETE",
		},
		{
			name:       "fails to delete project",
			want:       http.StatusInternalServerError,
			authHeader: adminAuthHeader,
			url:        "/projects/undeletableproject",
			method:     "DELETE",
		},
		{
			name:       "fails to delete project db entry",
			want:       http.StatusInternalServerError,
			authHeader: adminAuthHeader,
			url:        "/projects/somedeletedberror",
			method:     "DELETE",
		},
	}
	runTests(t, tests)
}

func TestGetProject(t *testing.T) {
	tests := []test{
		{
			name:       "cannot get project, when not admin",
			want:       http.StatusUnauthorized,
			authHeader: userAuthHeader,
			method:     "GET",
			url:        "/projects/project1",
		},
		{
			name:       "project exists, successful get project",
			want:       http.StatusOK,
			authHeader: adminAuthHeader,
			method:     "GET",
			url:        "/projects/project1",
		},
		{
			name:       "project does not exist",
			want:       http.StatusNotFound,
			authHeader: adminAuthHeader,
			method:     "GET",
			url:        "/projects/projectdoesnotexist",
		},
	}
	runTests(t, tests)
}

func TestCreateTarget(t *testing.T) {
	tests := []test{
		{
			name:       "can create target",
			req:        loadJSON(t, "TestCreateTarget/can_create_target_request.json"),
			want:       http.StatusOK,
			respFile:   "TestCreateTarget/can_create_target_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets",
			method:     "POST",
		},
		{
			name:       "fails to create target when not admin",
			req:        loadJSON(t, "TestCreateTarget/fails_to_create_target_when_not_admin_request.json"),
			want:       http.StatusUnauthorized,
			respFile:   "TestCreateTarget/fails_to_create_target_when_not_admin_response.json",
			authHeader: userAuthHeader,
			url:        "/projects/projectalreadyexists/targets",
			method:     "POST",
		},
		{
			name:       "fails to create target when using a bad auth header",
			req:        loadJSON(t, "TestCreateTarget/can_create_target_request.json"),
			want:       http.StatusUnauthorized,
			respFile:   "TestCreateTarget/fails_to_create_target_when_bad_auth_header_response.json",
			authHeader: invalidAuthHeader,
			url:        "/projects/projectalreadyexists/targets",
			method:     "POST",
		},
		{
			name:       "bad request",
			req:        loadJSON(t, "TestCreateTarget/bad_request.json"),
			want:       http.StatusBadRequest,
			respFile:   "TestCreateTarget/bad_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets",
			method:     "POST",
		},
		{
			name:       "target name cannot already exist",
			req:        loadJSON(t, "TestCreateTarget/target_name_cannot_already_exist_request.json"),
			want:       http.StatusBadRequest,
			respFile:   "TestCreateTarget/target_name_cannot_already_exist_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets",
			method:     "POST",
		},
		{
			name:       "project must exist",
			req:        loadJSON(t, "TestCreateTarget/project_must_exist_request.json"),
			want:       http.StatusBadRequest,
			respFile:   "TestCreateTarget/project_must_exist_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectdoesnotexist/targets",
			method:     "POST",
		},
	}
	runTests(t, tests)
}

func TestDeleteTarget(t *testing.T) {
	tests := []test{
		{
			name:       "fails to delete target when not admin",
			want:       http.StatusUnauthorized,
			authHeader: userAuthHeader,
			url:        "/projects/projectalreadyexists/targets/target1",
			method:     "DELETE",
		},
		{
			name:       "fails to delete target when using a bad auth header",
			want:       http.StatusUnauthorized,
			authHeader: invalidAuthHeader,
			url:        "/projects/projectalreadyexists/targets/target1",
			method:     "DELETE",
		},
		{
			name:       "can delete target",
			want:       http.StatusOK,
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets/target1",
			method:     "DELETE",
		},
		{
			name:       "target fails to delete",
			want:       http.StatusInternalServerError,
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets/undeletabletarget",
			method:     "DELETE",
		},
	}
	runTests(t, tests)
}

func TestUpdateTarget(t *testing.T) {
	tests := []test{
		{
			name:       "can update target",
			req:        loadJSON(t, "TestUpdateTarget/can_update_target_request.json"),
			want:       http.StatusOK,
			respFile:   "TestUpdateTarget/can_update_target_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets/TARGET_EXISTS",
			method:     "PATCH",
		},
		{
			name:       "fails to update target when not admin",
			req:        loadJSON(t, "TestUpdateTarget/fails_to_update_target_when_not_admin_request.json"),
			want:       http.StatusUnauthorized,
			respFile:   "TestUpdateTarget/fails_to_update_target_when_not_admin_response.json",
			authHeader: userAuthHeader,
			url:        "/projects/projectalreadyexists/targets/TARGET_EXISTS",
			method:     "PATCH",
		},
		{
			name:       "fails to update target when using a bad auth header",
			req:        loadJSON(t, "TestUpdateTarget/fails_to_update_target_when_not_admin_request.json"),
			want:       http.StatusUnauthorized,
			respFile:   "TestUpdateTarget/fails_to_update_target_when_bad_auth_header_response.json",
			authHeader: invalidAuthHeader,
			url:        "/projects/projectalreadyexists/targets/TARGET_EXISTS",
			method:     "PATCH",
		},
		{
			name:       "fails to update target credential_type",
			req:        loadJSON(t, "TestUpdateTarget/fails_to_update_credential_type_request.json"),
			want:       http.StatusBadRequest,
			respFile:   "TestUpdateTarget/fails_to_update_credential_type_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets/TARGET_EXISTS",
			method:     "PATCH",
		},
		{
			name:       "does not overwrite target name or type when in request",
			req:        loadJSON(t, "TestUpdateTarget/fails_to_update_target_name_request.json"),
			want:       http.StatusOK,
			respFile:   "TestUpdateTarget/fails_to_update_target_name_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets/TARGET_EXISTS",
			method:     "PATCH",
		},
		{
			name:       "target name must exist",
			req:        loadJSON(t, "TestUpdateTarget/target_name_must_exist_request.json"),
			want:       http.StatusNotFound,
			respFile:   "TestUpdateTarget/target_name_must_exist_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectalreadyexists/targets/INVALID_TARGET",
			method:     "PATCH",
		},
		{
			name:       "project must exist",
			req:        loadJSON(t, "TestUpdateTarget/project_must_exist_request.json"),
			want:       http.StatusNotFound,
			respFile:   "TestUpdateTarget/project_must_exist_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectdoesnotexist/targets/TARGET_EXISTS",
			method:     "PATCH",
		},
	}
	runTests(t, tests)
}

func TestCreateWorkflow(t *testing.T) {
	tests := []test{
		{
			name:       "can create workflows",
			req:        loadJSON(t, "TestCreateWorkflow/can_create_workflow_request.json"),
			want:       http.StatusOK,
			authHeader: userAuthHeader,
			respFile:   "TestCreateWorkflow/can_create_workflow_response.json",
			method:     "POST",
			url:        "/workflows",
		},
		// We test this specific validation as it's server side only.
		{
			name:       "framework must be valid",
			req:        loadJSON(t, "TestCreateWorkflow/framework_must_be_valid_request.json"),
			want:       http.StatusBadRequest,
			authHeader: userAuthHeader,
			respFile:   "TestCreateWorkflow/framework_must_be_valid_response.json",
			method:     "POST",
			url:        "/workflows",
		},
		// We test this specific validation as it's server side only.
		{
			name:       "type must be valid",
			respFile:   "TestCreateWorkflow/type_must_be_valid_response.json",
			req:        loadJSON(t, "TestCreateWorkflow/type_must_be_valid_request.json"),
			authHeader: userAuthHeader,
			want:       http.StatusBadRequest,
			method:     "POST",
			url:        "/workflows",
		},
		{
			name:       "project must exist",
			req:        loadJSON(t, "TestCreateWorkflow/project_must_exist.json"),
			authHeader: userAuthHeader,
			want:       http.StatusBadRequest,
			method:     "POST",
			url:        "/workflows",
		},
		{
			name:       "target must exist",
			req:        loadJSON(t, "TestCreateWorkflow/target_must_exist.json"),
			authHeader: userAuthHeader,
			want:       http.StatusBadRequest,
			method:     "POST",
			url:        "/workflows",
		},
		{
			name:       "cannot create workflow with bad auth header",
			req:        loadJSON(t, "TestCreateWorkflow/can_create_workflow_response.json"),
			want:       http.StatusUnauthorized,
			authHeader: invalidAuthHeader,
			method:     "POST",
			url:        "/workflows",
		},
		// TODO with admin credentials should fail
	}
	runTests(t, tests)
}

func TestCreateWorkflowFromGit(t *testing.T) {
	tests := []test{
		{
			name:       "can create workflows",
			req:        loadJSON(t, "TestCreateWorkflowFromGit/good_request.json"),
			want:       http.StatusOK,
			authHeader: userAuthHeader,
			respFile:   "TestCreateWorkflowFromGit/good_response.json",
			method:     "POST",
			url:        "/projects/project1/targets/target1/operations",
		},
		{
			name:       "bad request",
			req:        loadJSON(t, "TestCreateWorkflowFromGit/bad_request.json"),
			want:       http.StatusBadRequest,
			authHeader: userAuthHeader,
			respFile:   "TestCreateWorkflowFromGit/bad_response.json",
			method:     "POST",
			url:        "/projects/project1/targets/target1/operations",
		},
		// TODO with admin credentials should fail
	}
	runTests(t, tests)
}

func TestGetWorkflow(t *testing.T) {
	tests := []test{
		{
			name:       "workflow exists, successful get workflow",
			want:       http.StatusOK,
			authHeader: adminAuthHeader,
			method:     "GET",
			url:        "/workflows/WORKFLOW_ALREADY_EXISTS",
		},
		{
			name:       "workflow does not exist",
			want:       http.StatusInternalServerError,
			authHeader: adminAuthHeader,
			method:     "GET",
			url:        "/workflows/WORKFLOW_DOES_NOT_EXIST",
		},
	}
	runTests(t, tests)
}

func TestGetWorkflowLogs(t *testing.T) {
	tests := []test{
		{
			name:       "successful get workflow logs",
			want:       http.StatusOK,
			authHeader: adminAuthHeader,
			method:     "GET",
			url:        "/workflows/WORKFLOW_ALREADY_EXISTS/logs",
		},
		{
			name:       "workflow does not exist",
			want:       http.StatusInternalServerError,
			authHeader: adminAuthHeader,
			method:     "GET",
			url:        "/workflows/WORKFLOW_DOES_NOT_EXIST/logs",
		},
	}
	runTests(t, tests)
}

func TestListWorkflows(t *testing.T) {
	tests := []test{
		{
			name:       "can get workflows",
			want:       http.StatusOK,
			authHeader: userAuthHeader,
			method:     "GET",
			url:        "/projects/projects1/targets/target1/workflows",
		},
	}
	runTests(t, tests)
}

func TestDeleteToken(t *testing.T) {
	tests := []test{
		{
			name:       "fails to delete tokens when not admin",
			want:       http.StatusUnauthorized,
			respFile:   "TestDeleteToken/fails_to_delete_token_when_not_admin_response.json",
			authHeader: userAuthHeader,
			url:        "/projects/project/tokens/existingtoken",
			method:     "DELETE",
		},
		{
			name:       "can delete token",
			want:       http.StatusOK,
			respFile:   "TestDeleteToken/can_delete_token_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/project/tokens/existingtoken",
			method:     "DELETE",
			cpMock: &th.CredsProviderMock{
				GetProjectTokenFunc: func(s1 string, s2 string) (types.ProjectToken, error) {
					return types.ProjectToken{ID: "1234"}, nil
				},
				ProjectExistsFunc: func(s string) (bool, error) {
					return true, nil
				},
				DeleteProjectTokenFunc: func(p, t string) error {
					return nil
				},
			},
			dbMock: &th.DBClientMock{
				DeleteTokenEntryFunc: func(ctx context.Context, token string) error {
					return nil
				},
				ReadProjectEntryFunc: func(ctx context.Context, project string) (db.ProjectEntry, error) {
					return db.ProjectEntry{ProjectID: "project1"}, nil
				},
				ReadTokenEntryFunc: func(ctx context.Context, token string) (db.TokenEntry, error) {
					return db.TokenEntry{ProjectID: "project1", TokenID: "1234", CreatedAt: "2022-06-21T14:42:50.182037-07:00"}, nil
				},
			},
		},
		{
			name:       "project does not exist",
			want:       http.StatusNotFound,
			respFile:   "TestDeleteToken/project_does_not_exist_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectdoesnotexist/tokens/tokendoesnotexist",
			method:     "DELETE",
			cpMock: &th.CredsProviderMock{
				ProjectExistsFunc: func(s string) (bool, error) {
					return false, nil
				},
			},
		},
		{
			name:       "token does not exist",
			want:       http.StatusNotFound,
			respFile:   "TestDeleteToken/token_does_not_exist_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/project/tokens/tokendoesnotexist",
			method:     "DELETE",
			cpMock: &th.CredsProviderMock{
				GetProjectTokenFunc: func(s1 string, s2 string) (types.ProjectToken, error) {
					return types.ProjectToken{}, nil
				},
				ProjectExistsFunc: func(s string) (bool, error) {
					return true, nil
				},
			},
			dbMock: &th.DBClientMock{
				ReadProjectEntryFunc: func(ctx context.Context, project string) (db.ProjectEntry, error) {
					return db.ProjectEntry{ProjectID: "project1"}, nil
				},
			},
		},
		{
			name:       "token delete error",
			want:       http.StatusInternalServerError,
			respFile:   "TestDeleteToken/token_delete_error_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/project/tokens/deletetokenerror",
			method:     "DELETE",
			cpMock: &th.CredsProviderMock{
				DeleteProjectTokenFunc: func(s1, s2 string) error {
					return errors.New("error deleting token from Vault")
				},
				GetProjectTokenFunc: func(s1 string, s2 string) (types.ProjectToken, error) {
					return types.ProjectToken{ID: "1234"}, nil
				},
				ProjectExistsFunc: func(s string) (bool, error) {
					return true, nil
				},
			},
			dbMock: &th.DBClientMock{
				DeleteTokenEntryFunc: func(ctx context.Context, token string) error {
					return errors.New("error deleting entry from DB")
				},
				ReadProjectEntryFunc: func(ctx context.Context, project string) (db.ProjectEntry, error) {
					return db.ProjectEntry{ProjectID: "project1"}, nil
				},
				ReadTokenEntryFunc: func(ctx context.Context, token string) (db.TokenEntry, error) {
					return db.TokenEntry{ProjectID: "project1", TokenID: "1234", CreatedAt: "2022-06-21T14:42:50.182037-07:00"}, nil
				},
			},
		},
	}
	runTestsV2(t, tests)
}

func TestListTokens(t *testing.T) {
	tests := []test{
		{
			name:       "fails to list tokens when not admin",
			want:       http.StatusUnauthorized,
			respFile:   "TestListTokens/fails_to_list_tokens_when_not_admin_response.json",
			authHeader: userAuthHeader,
			url:        "/projects/undeletableprojecttargets/tokens",
			method:     "GET",
		},
		{
			name:       "can list tokens",
			want:       http.StatusOK,
			respFile:   "TestListTokens/can_list_tokens_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/undeletableprojecttargets/tokens",
			method:     "GET",
			dbMock: &th.DBClientMock{
				ReadProjectEntryFunc: func(ctx context.Context, p string) (db.ProjectEntry, error) {
					return db.ProjectEntry{ProjectID: "project1", Repository: "repo"}, nil
				},
				ListTokenEntriesFunc: func(ctx context.Context, project string) ([]db.TokenEntry, error) {
					return []db.TokenEntry{
						{
							CreatedAt: "2022-06-21T14:56:10.341066-07:00",
							TokenID:   "ghi789",
						},
						{
							CreatedAt: "2022-06-21T14:43:16.172896-07:00",
							TokenID:   "def456",
						},
						{
							CreatedAt: "2022-06-21T14:42:50.182037-07:00",
							TokenID:   "abc123",
						},
					}, nil
				},
			},
		},
		{
			name:       "project not found",
			want:       http.StatusNotFound,
			respFile:   "TestListTokens/project_not_found_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectdoesnotexist/tokens",
			method:     "GET",
			dbMock: &th.DBClientMock{
				ReadProjectEntryFunc: func(ctx context.Context, p string) (db.ProjectEntry, error) {
					return db.ProjectEntry{}, upper.ErrNoMoreRows
				},
			},
		},
		{
			name:       "no tokens",
			want:       http.StatusOK,
			respFile:   "TestListTokens/no_tokens_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectnotokens/tokens",
			method:     "GET",
			dbMock: &th.DBClientMock{
				ReadProjectEntryFunc: func(ctx context.Context, p string) (db.ProjectEntry, error) {
					return db.ProjectEntry{ProjectID: "abc123", Repository: "repo"}, nil
				},
				ListTokenEntriesFunc: func(ctx context.Context, project string) ([]db.TokenEntry, error) {
					return []db.TokenEntry{}, nil
				},
			},
		},
		{
			name:       "project read error",
			want:       http.StatusInternalServerError,
			respFile:   "TestListTokens/project_read_error_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectreaderror/tokens",
			method:     "GET",
			cpMock: &th.CredsProviderMock{
				ProjectExistsFunc: func(s string) (bool, error) {
					return false, errors.New("error retrieving project")
				},
			},
		},
		{
			name:       "list tokens read error",
			want:       http.StatusInternalServerError,
			respFile:   "TestListTokens/list_tokens_error_response.json",
			authHeader: adminAuthHeader,
			url:        "/projects/projectlisttokenserror/tokens",
			method:     "GET",
			dbMock: &th.DBClientMock{
				ReadProjectEntryFunc: func(ctx context.Context, project string) (db.ProjectEntry, error) {
					return db.ProjectEntry{ProjectID: "project1"}, nil
				},
				ListTokenEntriesFunc: func(ctx context.Context, project string) ([]db.TokenEntry, error) {
					return []db.TokenEntry{}, errors.New("error from DB")
				},
			},
		},
	}
	runTestsV2(t, tests)
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
			resp := executeRequest(tt.method, tt.url, serialize(tt.req), tt.authHeader)
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

				bodyStr := string(body)
				wantBodyStr := string(wantBody)

				if bodyStr == "" && wantBodyStr == "" {
					assert.Equal(t, bodyStr, wantBodyStr)
				} else {
					assert.JSONEq(t, bodyStr, wantBodyStr)
				}
			}
		})
	}
}

func runTestsV2(t *testing.T, tests []test) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := loadConfig(testConfigPath)
			if err != nil {
				panic(fmt.Sprintf("Unable to load config %s", err))
			}

			h := handler{
				logger:                 log.NewNopLogger(),
				newCredentialsProvider: newMockProvider,
				argo:                   mockWorkflowSvc{},
				argoCtx:                context.Background(),
				config:                 config,
				gitClient:              newMockGitClient(),
				env: env.Vars{
					AdminSecret: testPassword,
				},
				dbClient: newMockDB(),
			}

			if tt.dbMock != nil {
				h.dbClient = tt.dbMock
			}

			if tt.cpMock != nil {
				defaultCPFunc := func(a credentials.Authorization, env env.Vars, h http.Header, f credentials.VaultConfigFn, fn credentials.VaultSvcFn) (credentials.Provider, error) {
					// TODO: probably best to create a base/default CP mock struct here
					return tt.cpMock, nil
				}

				h.newCredentialsProvider = defaultCPFunc
			}

			resp := executeRequestWithHandler(h, tt.method, tt.url, serialize(tt.req), tt.authHeader)
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

				bodyStr := string(body)
				wantBodyStr := string(wantBody)

				// don't use assert.JSONEq for empty strings
				if bodyStr == "" && wantBodyStr == "" {
					assert.Equal(t, bodyStr, wantBodyStr)
				} else {
					assert.JSONEq(t, wantBodyStr, bodyStr)
				}
			}
		})
	}
}

// Execute a generic HTTP request, making sure to add the appropriate authorization header.
func executeRequest(method string, url string, body *bytes.Buffer, authHeader string) *http.Response {
	config, err := loadConfig(testConfigPath)
	if err != nil {
		panic(fmt.Sprintf("Unable to load config %s", err))
	}

	h := handler{
		logger:                 log.NewNopLogger(),
		newCredentialsProvider: newMockProvider,
		argo:                   mockWorkflowSvc{},
		argoCtx:                context.Background(),
		config:                 config,
		gitClient:              newMockGitClient(),
		env: env.Vars{
			AdminSecret: testPassword,
		},
		dbClient: newMockDB(),
	}

	var router = setupRouter(h)
	req, _ := http.NewRequest(method, url, body)

	req.Header.Add("Authorization", authHeader)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Result()
}

func executeRequestWithHandler(h handler, method string, url string, body *bytes.Buffer, authHeader string) *http.Response {
	var router = setupRouter(h)
	req, _ := http.NewRequest(method, url, serialize(body))

	req.Header.Add("Authorization", authHeader)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Result()
}

// loadFileBytes returns the contents of a file in the 'testdata' directory.
func loadFileBytes(filename string) ([]byte, error) {
	file := filepath.Join("test/testdata", filename)
	return os.ReadFile(file)
}

// loadJSON unmarshals a JSON file from the testdata directory into output.
func loadJSON(t *testing.T, filename string) (output interface{}) {
	data, err := loadFileBytes(filename)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", filename, err)
	}

	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to decode file %s: %v", filename, err)
	}
	return output
}
