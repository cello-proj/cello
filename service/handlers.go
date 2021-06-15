package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/argoproj-labs/argo-cloudops/service/internal/credentials"
	"github.com/argoproj-labs/argo-cloudops/service/internal/env"
	"github.com/argoproj-labs/argo-cloudops/service/internal/workflow"

	vault "github.com/hashicorp/vault/api"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/distribution/distribution/reference"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

// Create workflow request.
type createWorkflowRequest struct {
	Arguments            map[string][]string `yaml:"arguments" json:"arguments"`
	EnvironmentVariables map[string]string   `yaml:"environment_variables" json:"environment_variables"`
	Framework            string              `yaml:"framework" json:"framework"`
	Parameters           map[string]string   `yaml:"parameters" json:"parameters"`
	ProjectName          string              `yaml:"project_name" json:"project_name"`
	TargetName           string              `yaml:"target_name" json:"target_name"`
	Type                 string              `yaml:"type" json:"type"`
	WorkflowTemplateName string              `yaml:"workflow_template_name" json:"workflow_template_name"`
}

// Create workflow from git manifest request
type createGitWorkflowRequest struct {
	Repository string `json:"repository"`
	CommitHash string `json:"sha"`
	Path       string `json:"path"`
	Type       string `json:"type"`
}

// Represents a JWT token.
type token struct {
	Token string `json:"token"`
}

// Represents an error response.
type errorResponse struct {
	ErrorMessage string `json:"error_message"`
}

// Generates error response JSON.
func generateErrorResponseJSON(message string) string {
	er := errorResponse{
		ErrorMessage: message,
	}
	// TODO swallowing error since this is only internally ever passed message
	jsonData, _ := json.Marshal(er)
	return string(jsonData)
}

// Convenience method for checking if a string is alphanumeric.
var isStringAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString

// Vault does not allow for dashes
var isStringAlphaNumericUnderscore = regexp.MustCompile(`^([a-zA-Z])[a-zA-Z0-9_]*$`).MatchString

// HTTP handler
type handler struct {
	logger                 log.Logger
	newCredentialsProvider func(a credentials.Authorization, svc *vault.Client) (credentials.Provider, error)
	argo                   workflow.Workflow
	argoCtx                context.Context
	config                 *Config
	gitClient              gitClient
	env                    env.Vars
	newCredsProviderSvc    func(c credentials.VaultConfig, h http.Header) (*vault.Client, error)
	vaultConfig            credentials.VaultConfig
}

// Validates workflow parameters
func (h handler) validateWorkflowParameters(parameters map[string]string) error {
	if _, ok := parameters["execute_container_image_uri"]; !ok {
		return errors.New("parameters must include execute_container_image_uri")
	}

	if !h.isValidImageURI(parameters["execute_container_image_uri"]) {
		return errors.New("execute_container_image_uri must be a valid container uri")
	}

	if _, ok := parameters["pre_container_image_uri"]; ok {
		if !h.isValidImageURI(parameters["pre_container_image_uri"]) {
			return errors.New("pre_container_image_uri must be a valid container uri")
		}
	}

	return nil
}

// Service HealthCheck
func (h handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	level.Debug(h.logger).Log("message", "executing health check")
	vaultEndpoint := fmt.Sprintf("%s/v1/sys/health", os.Getenv("VAULT_ADDR"))

	// #nosec
	response, err := http.Get(vaultEndpoint)
	if err != nil {
		level.Error(h.logger).Log("message", fmt.Sprintf("received error connecting to vault endpoint %s", vaultEndpoint))
		h.errorResponse(w, "Health check failed", http.StatusServiceUnavailable)
		return
	}

	// We don't care about the body but need to read it all and close it
	// regardless.
	// https://golang.org/pkg/net/http/#Client.Do
	defer response.Body.Close()
	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		level.Error(h.logger).Log("message", "unable to read body; continuing", "error", err)
		// Continue on and handle the actual response code from Vault accordingly.
	}

	if response.StatusCode != 200 && response.StatusCode != 429 {
		level.Error(h.logger).Log("message", fmt.Sprintf("received code %d which is not 200 (initialized, unsealed, and active) or 429 (unsealed and standby) when connecting to vault endpoint %s", response.StatusCode, vaultEndpoint))
		h.errorResponse(w, "Health check failed", http.StatusServiceUnavailable)
	} else {
		fmt.Fprintln(w, "Health check succeeded")
	}
}

// Lists workflows
func (h handler) listWorkflows(w http.ResponseWriter, r *http.Request) {
	// TODO authenticate user can list this workflow once auth figured out
	// TODO fail if project / target does not exist or are not valid format
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	targetName := vars["targetName"]

	l := h.requestLogger(r, "op", "list-workflows", "project", projectName, "target", targetName)

	level.Debug(l).Log("message", "listing workflows")
	workflowIDs, err := h.argo.List(h.argoCtx)
	if err != nil {
		level.Error(l).Log("message", "error listing workflows", "error", err)
		h.errorResponse(w, "error listing workflows", http.StatusBadRequest)
		return
	}

	// Only return workflows the target project / target
	var workflows []workflow.Status
	prefix := fmt.Sprintf("%s-%s", projectName, targetName)
	for _, workflowID := range workflowIDs {
		if strings.HasPrefix(workflowID, prefix) {
			workflow, err := h.argo.Status(h.argoCtx, workflowID)
			if err != nil {
				level.Error(l).Log("message", "error retrieving workflows", "error", err)
				h.errorResponse(w, "error retrieving workflows", http.StatusBadRequest)
				return
			}
			workflows = append(workflows, *workflow)
		}
	}

	jsonData, err := json.Marshal(workflows)
	if err != nil {
		level.Error(l).Log("message", "error serializing workflow IDs", "error", err)
		h.errorResponse(w, "error serializing workflow IDs", http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, string(jsonData))
}

// Creates workflow init params by pulling manifest from given git repo, commit sha, and code path
func (h handler) loadCreateWorkflowRequestFromGit(repository, commitHash, path string) (createWorkflowRequest, error) {
	level.Debug(h.logger).Log("message", fmt.Sprintf("retrieving manifest from repository %s at sha %s with path %s", repository, commitHash, path))
	fileContents, err := h.gitClient.CheckoutFileFromRepository(repository, commitHash, path)
	if err != nil {
		return createWorkflowRequest{}, err
	}

	var cwr createWorkflowRequest
	err = yaml.Unmarshal(fileContents, &cwr)
	return cwr, err
}

func (h handler) createWorkflowFromGit(w http.ResponseWriter, r *http.Request) {
	l := h.requestLogger(r, "op", "create-workflow-from-git")

	ctx := r.Context()

	level.Debug(l).Log("message", "reading request body")
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		level.Error(l).Log("message", "error reading authorization data", "error", err)
		h.errorResponse(w, "error reading authorization data", http.StatusInternalServerError)
		return
	}

	var cgwr createGitWorkflowRequest
	err = json.Unmarshal(reqBody, &cgwr)
	if err != nil {
		level.Error(l).Log("message", "error deserializing request body", "error", err)
		h.errorResponse(w, "error deserializing request body", http.StatusBadRequest)
		return
	}

	ah := r.Header.Get("Authorization")
	a, err := credentials.NewAuthorization(ah)
	if err != nil {
		h.errorResponse(w, "error authorizing", http.StatusUnauthorized)
		return
	}

	cwr, err := h.loadCreateWorkflowRequestFromGit(cgwr.Repository, cgwr.CommitHash, cgwr.Path)
	if err != nil {
		level.Error(l).Log("message", "error loading workflow data from git", "error", err)
		h.errorResponse(w, "error loading workflow data from git", http.StatusBadRequest)
		return
	}

	log.With(l, "project", cwr.ProjectName, "target", cwr.TargetName, "framework", cwr.Framework, "type", cwr.Type, "workflow-template", cwr.WorkflowTemplateName)

	level.Debug(l).Log("message", "creating workflow")
	cwr.Type = cgwr.Type
	h.createWorkflowFromRequest(ctx, w, r, a, cwr, l)
}

// Creates a workflow
func (h handler) createWorkflow(w http.ResponseWriter, r *http.Request) {
	l := h.requestLogger(r, "op", "create-workflow")

	ctx := r.Context()

	level.Debug(l).Log("message", "reading request body")
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		level.Error(l).Log("message", "error reading authorization data", "error", err)
		h.errorResponse(w, "error reading authorization data", http.StatusInternalServerError)
		return
	}

	var cwr createWorkflowRequest
	err = json.Unmarshal(reqBody, &cwr)
	if err != nil {
		level.Error(l).Log("message", "error deserializing workflow data", "error", err)
		h.errorResponse(w, "error deserializing workflow data", http.StatusBadRequest)
		return
	}

	log.With(l, "project", cwr.ProjectName, "target", cwr.TargetName, "framework", cwr.Framework, "type", cwr.Type, "workflow-template", cwr.WorkflowTemplateName)

	ah := r.Header.Get("Authorization")
	a, err := credentials.NewAuthorization(ah)
	if err != nil {
		h.errorResponse(w, "error authorizing", http.StatusUnauthorized)
		return
	}

	level.Debug(l).Log("message", "creating workflow")
	h.createWorkflowFromRequest(ctx, w, r, a, cwr, l)
}

// Creates a workflow
// Context is not currently used as Argo has its own and Vault doesn't
// currently support it.
func (h handler) createWorkflowFromRequest(_ context.Context, w http.ResponseWriter, r *http.Request, a *credentials.Authorization, cwr createWorkflowRequest, l log.Logger) {
	level.Debug(l).Log("message", "validating workflow parameters")
	if err := h.validateWorkflowParameters(cwr.Parameters); err != nil {
		level.Error(l).Log("message", "error in parameters", "error", err)
		h.errorResponse(w, "error in parameters", http.StatusInternalServerError)
		return
	}

	frameworks := h.config.listFrameworks()

	if !stringInSlice(frameworks, cwr.Framework) {
		level.Error(l).Log("error", "unknown framework")
		h.errorResponse(w, "unknown framework", http.StatusBadRequest)
		return
	}

	types, err := h.config.listTypes(cwr.Framework)
	if err != nil {
		level.Error(l).Log("message", "error reading types from config", "error", err)
		h.errorResponse(w, "error reading types from config", http.StatusBadRequest)
		return
	}

	if !stringInSlice(types, cwr.Type) {
		level.Error(l).Log("error", "unknown type")
		h.errorResponse(w, "unknown type", http.StatusBadRequest)
		return
	}

	// TODO long term, we should evaluate if hard coding in code is the right approach to
	// specifying different argument types vs allowing dynmaic specification and
	// interpolation in service/config.yaml
	for k := range cwr.Arguments {
		if k != "init" && k != "execute" {
			level.Error(l).Log("message", "arguments must be init or execute", "error", err)
			h.errorResponse(w, "arguments must be init or execute", http.StatusBadRequest)
			return
		}
	}

	isValidProjectName, err := h.validateProjectName(cwr.ProjectName, w)
	if !isValidProjectName {
		level.Error(l).Log("error", err)
		return
	}

	isValidTargetName, err := h.validateTargetName(cwr.TargetName, w)
	if !isValidTargetName {
		level.Error(l).Log("error", err)
		return
	}

	// TODO: Fix type must be specified valication and add test
	//	h.errorResponse(w, "type must be specified", http.StatusBadRequest)
	//	return
	//}

	workflowFrom := fmt.Sprintf("workflowtemplate/%s", cwr.WorkflowTemplateName)
	executeContainerImageURI := cwr.Parameters["execute_container_image_uri"]
	environmentVariablesString := generateEnvVariablesString(cwr.EnvironmentVariables)

	level.Debug(l).Log("message", "generating command to execute")
	commandDefinition, err := h.config.getCommandDefinition(cwr.Framework, cwr.Type)
	if err != nil {
		level.Error(l).Log("message", "unable to get command definition", "error", err)
		h.errorResponse(w, "unable to get command definition", http.StatusBadRequest)
		return
	}
	executeCommand, err := generateExecuteCommand(commandDefinition, environmentVariablesString, cwr.Arguments)
	if err != nil {
		level.Error(l).Log("message", "unable to generate command", "error", err)
		h.errorResponse(w, "unable to generate command", http.StatusBadRequest)
		return
	}

	cpSvc, err := h.newCredsProviderSvc(h.vaultConfig, r.Header)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider service", "error", err)
		h.errorResponse(w, "error creating credentials provider service", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating new credentials provider")
	cp, err := h.newCredentialsProvider(*a, cpSvc)
	if err != nil {
		level.Error(l).Log("message", "bad or unknown credentials provider", "error", err)
		h.errorResponse(w, "bad or unknown credentials provider", http.StatusInternalServerError)
		return
	}

	level.Debug(l).Log("message", "getting credentials provider token")
	credentialsToken, err := cp.GetToken()
	if err != nil {
		level.Error(l).Log("message", "error getting credentials provider token", "error", err)
		h.errorResponse(w, "error getting credentials provider token", http.StatusInternalServerError)
		return
	}

	projectExists, err := cp.ProjectExists(cwr.ProjectName)
	if err != nil {
		level.Error(l).Log("message", "error checking project", "error", err)
		h.errorResponse(w, "error checking project", http.StatusInternalServerError)
		return
	}

	if !projectExists {
		level.Error(l).Log("message", "project does not exist", "error", err)
		h.errorResponse(w, "project does not exist", http.StatusBadRequest)
		return
	}

	// TODO: handle error when implemented
	//targetExists, _ := cp.targetExists(cwr.TargetName)
	//if !targetExists {
	//	h.errorResponse(w, "target must already exist", http.StatusBadRequest)
	//	return
	//}
	level.Debug(l).Log("message", "creating workflow parameters")
	parameters := workflow.NewParameters(environmentVariablesString, executeCommand, executeContainerImageURI, cwr.TargetName, cwr.ProjectName, cwr.Parameters, credentialsToken)

	level.Debug(l).Log("message", "creating workflow")
	workflowName, err := h.argo.Submit(h.argoCtx, workflowFrom, parameters)
	if err != nil {
		level.Error(l).Log("message", "error creating workflow", "error", err)
		h.errorResponse(w, "error creating workflow", http.StatusInternalServerError)
		return
	}

	l = log.With(l, "workflow", workflowName)
	level.Debug(l).Log("message", "workflow created")
	tokenHead := credentialsToken[0:8]

	level.Info(l).Log("message", fmt.Sprintf("Received token '%s...'", tokenHead))
	var cwresp workflow.CreateWorkflowResponse
	cwresp.WorkflowName = workflowName
	jsonData, err := json.Marshal(cwresp)
	if err != nil {
		level.Error(l).Log("message", "error serializing workflow response", "error", err)
		h.errorResponse(w, "error serializing workflow response", http.StatusBadRequest)
		return
	}
	fmt.Fprintln(w, string(jsonData))
}

// Gets a workflow
func (h handler) getWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowName := vars["workflowName"]
	// TODO: Workflow name must include -
	// need to update validation
	//if !h.validateWorkflowName(workflowName, w) {
	//	return
	//}
	l := h.requestLogger(r, "op", "get-workflow", "workflow", workflowName)

	level.Debug(l).Log("message", "getting workflow status")
	status, err := h.argo.Status(h.argoCtx, workflowName)
	if err != nil {
		level.Error(l).Log("message", "error getting workflow", "error", err)
		h.errorResponse(w, "error getting workflow", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "decoding get workflow response")
	jsonData, err := json.Marshal(status) // TODO handle error in http resp
	if err != nil {
		level.Error(l).Log("message", "error serializing workflow", "error", err)
		h.errorResponse(w, "error serializing workflow", http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, string(jsonData))
}

// Gets a workflow
func (h handler) getTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	targetName := vars["targetName"]

	l := h.requestLogger(r, "op", "get-target", "project", projectName, "target", targetName)

	level.Debug(l).Log("message", "authorizing get target permissions")
	ah := r.Header.Get("Authorization")
	a, err := credentials.NewAuthorization(ah) // todo add validation
	if err != nil {
		level.Error(l).Log("message", "error authorizing using Authorization header", "error", err)
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized)
		return
	}

	level.Debug(l).Log("message", "validating authorized admin")
	if !a.AuthorizedAdmin(h.env.AdminSecret) {
		level.Error(l).Log("message", "must be an authorized admin")
		h.errorResponse(w, "must be an authorized admin", http.StatusUnauthorized)
		return
	}

	cpSvc, err := h.newCredsProviderSvc(h.vaultConfig, r.Header)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider service", "error", err)
		h.errorResponse(w, "error creating credentials provider service", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a, cpSvc)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider", "error", err)
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "getting target information")
	targetInfo, err := cp.GetTarget(projectName, targetName)
	if err != nil {
		level.Error(l).Log("message", "error retrieving target information", "error", err)
		h.errorResponse(w, "error retrieving target information", http.StatusBadRequest)
		return
	}

	jsonResult, err := json.Marshal(targetInfo)
	if err != nil {
		level.Error(l).Log("message", "error serializing json target data", "error", err)
		h.errorResponse(w, "error serializing json target data", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(jsonResult))
}

// Returns the logs for a workflow
func (h handler) getWorkflowLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowName := vars["workflowName"]
	// TODO: Workflow name must include -
	// need to update validation
	//if !h.validateWorkflowName(workflowName, w) {
	//	return
	//}
	l := h.requestLogger(r, "op", "get-workflow-logs", "workflow", workflowName)

	level.Debug(l).Log("message", "retrieving workflow logs")
	argoWorkflowLogs, err := h.argo.Logs(h.argoCtx, workflowName)
	if err != nil {
		level.Error(l).Log("message", "error getting workflow logs", "error", err)
		h.errorResponse(w, "error getting workflow logs", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(argoWorkflowLogs)
	if err != nil {
		level.Error(l).Log("message", "error serializing workflow logs", "error", err)
		h.errorResponse(w, "error serializing workflow logs", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(jsonData))
}

// Streams workflow logs
func (h handler) getWorkflowLogStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	vars := mux.Vars(r)
	workflowName := vars["workflowName"]
	// TODO: Workflow name must include -
	// need to update validation
	//if !h.validateWorkflowName(workflowName, w) {
	//	return
	//}
	l := h.requestLogger(r, "op", "get-workflow-log-stream", "workflow", workflowName)

	level.Debug(l).Log("message", "retrieving workflow logs", "workflow", workflowName)
	err := h.argo.LogStream(h.argoCtx, workflowName, w)
	if err != nil {
		level.Error(l).Log("message", "error getting workflow logstream", "error", err)
		h.errorResponse(w, "error getting workflow logs", http.StatusBadRequest)
		return
	}
}

// Returns a new token
func newArgoCloudOpsToken(provider, key, secret string) *token {
	return &token{
		Token: fmt.Sprintf("%s:%s:%s", provider, key, secret),
	}
}

// Creates a project
func (h handler) createProject(w http.ResponseWriter, r *http.Request) {
	l := h.requestLogger(r, "op", "create-project")

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		level.Error(l).Log("message", "error reading request body", "error", err)
		h.errorResponse(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	var capp credentials.CreateProjectRequest
	err = json.Unmarshal(reqBody, &capp)
	if err != nil {
		level.Error(l).Log("message", "error parsing json", "error", err)
		h.errorResponse(w, "error parsing json", http.StatusInternalServerError)
		return
	}

	l = log.With(l, "project", capp.Name)

	level.Debug(l).Log("message", "authorizing project creation")
	ah := r.Header.Get("Authorization")
	a, err := credentials.NewAuthorization(ah) // todo add validation
	if err != nil {
		level.Error(l).Log("message", "error authorizing using Authorization header", "error", err)
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized)
		return
	}

	level.Debug(l).Log("message", "validating authorized admin")
	if !a.AuthorizedAdmin(h.env.AdminSecret) {
		level.Error(l).Log("message", "must be an authorized admin")
		h.errorResponse(w, "must be an authorized admin", http.StatusUnauthorized)
		return
	}

	cpSvc, err := h.newCredsProviderSvc(h.vaultConfig, r.Header)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider service", "error", err)
		h.errorResponse(w, "error creating credentials provider service", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a, cpSvc)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider", "error", err)
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest)
		return
	}

	isValidProjectName, err := h.validateProjectName(capp.Name, w)
	if !isValidProjectName {
		level.Error(l).Log("message", err)
		return
	}

	projectExists, err := cp.ProjectExists(capp.Name)
	if err != nil {
		level.Error(l).Log("message", "error checking project", "error", err)
		h.errorResponse(w, "error checking project", http.StatusInternalServerError)
		return
	}

	if projectExists {
		level.Error(l).Log("error", "project already exists")
		h.errorResponse(w, "project already exists", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating project")
	role, secret, err := cp.CreateProject(capp.Name)
	if err != nil {
		level.Error(l).Log("message", "error creating project", "error", err)
		h.errorResponse(w, "error creating project", http.StatusInternalServerError)
		return
	}

	level.Debug(l).Log("message", "retrieving Argo CloudOps token")
	t := newArgoCloudOpsToken("vault", role, secret)
	jsonResult, err := json.Marshal(t)
	if err != nil {
		level.Error(l).Log("message", "error serializing token", "error", err)
		h.errorResponse(w, "error serializing token", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(jsonResult))
}

// Get a project
func (h handler) getProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	l := h.requestLogger(r, "op", "get-project", "project", projectName)

	level.Debug(l).Log("message", "authorizing get project")
	ah := r.Header.Get("Authorization")
	a, err := credentials.NewAuthorization(ah) // todo add validation
	if err != nil {
		level.Error(l).Log("message", "error authorizing using Authorization token", "error", err)
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized)
		return
	}

	level.Debug(l).Log("message", "validating authorized admin")
	if !a.AuthorizedAdmin(h.env.AdminSecret) {
		level.Error(l).Log("message", "must be an authorized admin")
		h.errorResponse(w, "must be an authorized admin", http.StatusUnauthorized)
		return
	}

	cpSvc, err := h.newCredsProviderSvc(h.vaultConfig, r.Header)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider service", "error", err)
		h.errorResponse(w, "error creating credentials provider service", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a, cpSvc)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider", "error", err)
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "getting project")
	jsonResult, err := cp.GetProject(projectName)
	if err != nil {
		level.Error(l).Log("message", "error retrieving project", "error", err)
		h.errorResponse(w, "error retrieving project", http.StatusNotFound)
		return
	}
	fmt.Fprint(w, jsonResult)
}

// Delete a project
func (h handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	l := h.requestLogger(r, "op", "delete-project", "project", projectName)

	level.Debug(l).Log("message", "authorizing delete project")
	ah := r.Header.Get("Authorization")
	a, err := credentials.NewAuthorization(ah) // todo add validation
	if err != nil {
		level.Error(l).Log("message", "error authorizing using Authorization header", "error", err)
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized)
		return
	}

	level.Debug(l).Log("message", "validating authorized admin")
	if !a.AuthorizedAdmin(h.env.AdminSecret) {
		level.Error(l).Log("message", "must be an authorized admin")
		h.errorResponse(w, "must be an authorized admin", http.StatusUnauthorized)
		return
	}

	cpSvc, err := h.newCredsProviderSvc(h.vaultConfig, r.Header)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider service", "error", err)
		h.errorResponse(w, "error creating credentials provider service", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a, cpSvc)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider", "error", err)
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "checking if project exists")
	projectExists, err := cp.ProjectExists(projectName)
	if err != nil {
		level.Error(l).Log("message", "error checking project", "error", err)
		h.errorResponse(w, "error checking project", http.StatusInternalServerError)
		return
	}

	if !projectExists {
		level.Debug(l).Log("message", "no action required because project does not exist")
		return
	}

	level.Debug(l).Log("message", "getting all targets in project")
	targets, err := cp.ListTargets(projectName)
	if err != nil {
		level.Error(l).Log("message", "error getting all targets", "error", err)
		h.errorResponse(w, "error getting all targets", http.StatusInternalServerError)
		return
	}

	if len(targets) > 0 {
		level.Error(l).Log("error", "project has existing targets, not deleting")
		h.errorResponse(w, "project has existing targets, not deleting", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "deleting project")
	err = cp.DeleteProject(projectName)
	if err != nil {
		level.Error(l).Log("message", "error deleting project", "error", err)
		h.errorResponse(w, "error deleting project", http.StatusBadRequest)
		return
	}
}

// Creates a target
func (h handler) createTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	ah := r.Header.Get("Authorization")

	l := h.requestLogger(r, "op", "create-target", "project", projectName)

	level.Debug(l).Log("message", "reading request body")
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		level.Error(l).Log("message", "error reading request body", "error", err)
		h.errorResponse(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	var ctr credentials.CreateTargetRequest
	err = json.Unmarshal(reqBody, &ctr)
	if err != nil {
		level.Error(l).Log("message", "error parsing request body to target data", "error", err)
		h.errorResponse(w, "error parsing request body to target data", http.StatusBadRequest)
		return
	}

	l = log.With(l, "target", ctr.Name)

	level.Debug(l).Log("message", "authorizing target creation")
	a, err := credentials.NewAuthorization(ah)
	if err != nil {
		level.Error(l).Log("message", "error authorizing using Authorization header", "error", err)
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized)
		return
	}

	level.Debug(l).Log("message", "validating authorized admin")

	if !a.AuthorizedAdmin(h.env.AdminSecret) {
		level.Error(l).Log("message", "must be an authorized admin")
		h.errorResponse(w, "must be an authorized admin", http.StatusUnauthorized)
		return
	}

	cpSvc, err := h.newCredsProviderSvc(h.vaultConfig, r.Header)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider service", "error", err)
		h.errorResponse(w, "error creating credentials provider service", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a, cpSvc)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider", "error", err)
		h.errorResponse(w, "error creating credentials provider", http.StatusInternalServerError)
		return
	}

	level.Debug(l).Log("message", "validating target name")
	isValidTargetName, err := h.validateTargetName(ctr.Name, w)
	if !isValidTargetName {
		level.Error(l).Log("error", err)
		return
	}

	if ctr.Type != "aws_account" {
		level.Error(l).Log("error", "type must be aws_account")
		h.errorResponse(w, "type must be aws_account", http.StatusBadRequest)
		return
	}

	if len(ctr.Properties.PolicyArns) > 5 {
		level.Error(l).Log("error", "policy arns list length cannot be greater than 5")
		h.errorResponse(w, "policy arns list length cannot be greater than 5", http.StatusBadRequest)
		return
	}

	for _, policyArn := range ctr.Properties.PolicyArns {
		if !arn.IsARN(policyArn) {
			level.Error(l).Log("message", fmt.Sprintf("policy arn %s must be a valid arn", policyArn))
			h.errorResponse(w, fmt.Sprintf("policy arn %s must be a valid arn", policyArn), http.StatusBadRequest)
			return
		}
	}

	if !arn.IsARN(ctr.Properties.RoleArn) {
		level.Error(l).Log("message", fmt.Sprintf("role arn %s must be a valid arn", ctr.Properties.RoleArn))
		h.errorResponse(w, fmt.Sprintf("role arn %s must be a valid arn", ctr.Properties.RoleArn), http.StatusBadRequest)
		return
	}

	targetExists, _ := cp.TargetExists(ctr.Name)
	// TODO: handle error when implemented
	if targetExists {
		level.Error(l).Log("message", "target name must not already exist")
		h.errorResponse(w, "target name must not already exist", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating target")
	err = cp.CreateTarget(projectName, ctr)
	if err != nil {
		level.Error(l).Log("message", "error creating target", "error", err)
		h.errorResponse(w, "error creating target", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "{}")
}

// Deletes a target
func (h handler) deleteTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	targetName := vars["targetName"]

	l := h.requestLogger(r, "op", "delete-target", "project", projectName, "target", targetName)

	level.Debug(l).Log("message", "authorizing delete target permissions")
	ah := r.Header.Get("Authorization")
	a, err := credentials.NewAuthorization(ah) // todo add validation
	if err != nil {
		level.Error(l).Log("message", "error authorizing using Authorization header", "error", err)
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized)
		return
	}

	level.Debug(l).Log("message", "validating authorized admin")

	if !a.AuthorizedAdmin(h.env.AdminSecret) {
		level.Error(l).Log("message", "must be an authorized admin")
		h.errorResponse(w, "must be an authorized admin", http.StatusUnauthorized)
		return
	}

	cpSvc, err := h.newCredsProviderSvc(h.vaultConfig, r.Header)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider service", "error", err)
		h.errorResponse(w, "error creating credentials provider service", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a, cpSvc)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider", "error", err)
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "deleting target")
	err = cp.DeleteTarget(projectName, targetName)
	if err != nil {
		level.Error(l).Log("message", "error deleting target", "error", err)
		h.errorResponse(w, "error deleting target", http.StatusBadRequest)
		return
	}
}

// Lists the targets for a project
func (h handler) listTargets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	ah := r.Header.Get("Authorization")

	l := h.requestLogger(r, "op", "list-targets", "project", projectName)

	level.Debug(l).Log("message", "authorizing target list retrieval")
	a, err := credentials.NewAuthorization(ah)
	if err != nil {
		level.Error(l).Log("message", "error authorizing using Authorization header", "error", err)
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized)
		return
	}

	level.Debug(l).Log("message", "validating authorized admin")

	if !a.AuthorizedAdmin(h.env.AdminSecret) {
		level.Error(l).Log("message", "must be an authorized admin")
		h.errorResponse(w, "must be an authorized admin", http.StatusUnauthorized)
		return
	}

	cpSvc, err := h.newCredsProviderSvc(h.vaultConfig, r.Header)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider service", "error", err)
		h.errorResponse(w, "error creating credentials provider service", http.StatusBadRequest)
		return
	}

	level.Debug(l).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a, cpSvc)
	if err != nil {
		level.Error(l).Log("message", "error creating credentials provider", "error", err)
		h.errorResponse(w, "error creating credentials provider", http.StatusInternalServerError)
		return
	}

	targets, err := cp.ListTargets(projectName)
	if err != nil {
		level.Error(l).Log("message", "error listing targets", "error", err)
		h.errorResponse(w, "error listing targets", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(targets)
	if err != nil {
		level.Error(l).Log("message", "error serializing targets", "error", err)
	}

	fmt.Fprint(w, string(data))
}

// Convenience method that writes a failure response in a standard manner
func (h handler) errorResponse(w http.ResponseWriter, message string, httpStatus int) {
	r := generateErrorResponseJSON(message)
	w.WriteHeader(httpStatus)
	fmt.Fprint(w, r)
}

// Validates a project name
func (h handler) validateProjectName(projectName string, w http.ResponseWriter) (bool, error) {
	if len(projectName) < 4 {
		h.errorResponse(w, "project name must be greater than 3 characters", http.StatusBadRequest)
		return false, errors.New("project name must be greater than 3 characters")
	}

	if len(projectName) > 32 {
		h.errorResponse(w, "project name must be less than 32 characters", http.StatusBadRequest)
		return false, errors.New("project name must be less than 32 characters")
	}

	if !isStringAlphaNumeric(projectName) {
		h.errorResponse(w, "project name must be alpha-numeric", http.StatusBadRequest)
		return false, errors.New("project name must be alpha-numeric")
	}

	return true, nil
}

// Validates a target name
func (h handler) validateTargetName(targetName string, w http.ResponseWriter) (bool, error) {
	if len(targetName) < 4 {
		h.errorResponse(w, "target name must be greater than 3 characters", http.StatusBadRequest)
		return false, errors.New("target name must be greater than 3 characters")
	}

	if len(targetName) > 32 {
		h.errorResponse(w, "target name must be less than 32 characters", http.StatusBadRequest)
		return false, errors.New("target name must be less than 32 characters")
	}

	if !isStringAlphaNumericUnderscore(targetName) {
		h.errorResponse(w, "target name must be alpha-numeric with underscores", http.StatusBadRequest)
		return false, errors.New("target name must be alpha-numeric with underscores")
	}

	return true, nil
}

// Returns true, if the image uri is a valid container image uri
func (h handler) isValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
}

func generateEnvVariablesString(environmentVariables map[string]string) string {
	if len(environmentVariables) == 0 {
		return ""
	}

	r := "env"
	for k, v := range environmentVariables {
		tmp := r + fmt.Sprintf(" %s=%s", k, v)
		r = tmp
	}
	return r
}

func (h handler) requestLogger(r *http.Request, fields ...string) log.Logger {
	return log.With(h.logger, "txid", r.Header.Get(txIDHeader), fields)
}
