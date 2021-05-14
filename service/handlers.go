package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/distribution/distribution/reference"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	vault "github.com/hashicorp/vault/api"
)

// Create workflow request
type createWorkflowRequest struct {
	Arguments            map[string][]string `json:"arguments"`
	EnvironmentVariables map[string]string   `json:"environment_variables"`
	Framework            string              `json:"framework"`
	Parameters           map[string]string   `json:"parameters"`
	ProjectName          string              `json:"project_name"`
	TargetName           string              `json:"target_name"`
	Type                 string              `json:"type"`
	WorkflowTemplateName string              `json:"workflow_template_name"`
}

// Represents a JWT token
type token struct {
	Token string `json:"token"`
}

// Represents an error response
type errorResponse struct {
	ErrorMessage string `json:"error_message"`
}

// Generates error response JSON
func generateErrorResponseJSON(message string) string {
	er := errorResponse{
		ErrorMessage: message,
	}
	// TODO swallowing error since this is only internally ever passed message
	jsonData, _ := json.Marshal(er)
	return string(jsonData)
}

// Convenience method for checking if a string is alphanumeric
var isStringAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9_]*$`).MatchString

// Vault does not allow for dashes
var isStringAlphaNumericUnderscore = regexp.MustCompile(`^([a-zA-Z])[a-zA-Z0-9_]*$`).MatchString

// Represents a user's authorization token
type Authorization struct {
	Provider string
	Key      string
	Secret   string
}

// Authorization function for token requests
// this is separate from admin functions which use the admin env var
func newAuthorization(authorizationHeader string) (*Authorization, error) {
	var a Authorization
	auth := strings.SplitN(authorizationHeader, ":", 3)
	for _, i := range auth {
		if i == "" {
			return nil, fmt.Errorf("invalid authorization header provided")
		}
	}
	if len(auth) < 3 {
		return nil, fmt.Errorf("invalid authorization header provided")
	}
	a.Provider = auth[0]
	a.Key = auth[1]
	a.Secret = auth[2]
	return &a, nil
}

// Returns true, if the user is an admin
func (a Authorization) isAdmin() bool {
	return a.Key == "admin"
}

// Returns true, if the user is an authorized admin
func (a Authorization) authorizedAdmin() bool {
	return a.isAdmin() && a.Secret == adminSecret()
}

// HTTP handler
type handler struct {
	logger                 log.Logger
	logLevel               string
	newCredentialsProvider func(a Authorization) (credentialsProvider, error)
	argo                   Workflow
	config                 *Config
}

// Returns a new vaultCredentialsProvider
func newVaultProvider(svc *vault.Client) func(a Authorization) (credentialsProvider, error) {
	return func(a Authorization) (credentialsProvider, error) {
		return &vaultCredentialsProvider{
			VaultSvc: svc,
			RoleID:   a.Key,
			SecretID: a.Secret,
		}, nil
	}
}

// Validates workflow parameters
func (h handler) validateWorkflowParameters(parameters map[string]string) error {
	if _, ok := parameters["execute_container_image_uri"]; !ok {
		return errors.New("parameters must include execute_container_image_uri")
	}

	if !h.isValidImageUri(parameters["execute_container_image_uri"]) {
		return errors.New("execute_container_image_uri must be a valid container uri")
	}

	if _, ok := parameters["pre_container_image_uri"]; ok {
		if !h.isValidImageUri(parameters["pre_container_image_uri"]) {
			return errors.New("pre_container_image_uri must be a valid container uri")
		}
	}

	return nil
}

// Service HealthCheck
func (h handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	return
}

// Lists workflows
func (h handler) listWorkflows(w http.ResponseWriter, r *http.Request) {
	// TODO authenticate user can list this workflow once auth figured out
	// TODO fail if project / target does not exist or are not valid format
	level.Debug(h.logger).Log("message", "listing workflows")
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	targetName := vars["targetName"]

	workflowIDs, err := h.argo.ListWorkflows()
	if err != nil {
		h.errorResponse(w, "error listing workflows", http.StatusBadRequest, err)
		return
	}

	// Only return workflows the target project / target
	filteredWorkflowIDs := []string{}
	var workflows []workflowStatus
	prefix := fmt.Sprintf("%s-%s", projectName, targetName)
	for _, workflowID := range workflowIDs {
		if strings.HasPrefix(workflowID, prefix) {
			filteredWorkflowIDs = append(filteredWorkflowIDs, workflowID)
			workflow, err := h.argo.GetStatus(workflowID)
			if err != nil {
				h.errorResponse(w, "error retrieving workflows", http.StatusBadRequest, err)
				return
			}
			workflows = append(workflows, *workflow)
		}
	}

	jsonData, err := json.Marshal(workflows)
	if err != nil {
		h.errorResponse(w, "error serializing workflow ids", http.StatusBadRequest, err)
		return
	}

	fmt.Fprintln(w, string(jsonData))
}

// Creates a workflow
func (h handler) createWorkflow(w http.ResponseWriter, r *http.Request) {
	level.Debug(h.logger).Log("message", "creating workflow")
	ah := r.Header.Get("Authorization")
	a, err := newAuthorization(ah)
	if err != nil {
		h.errorResponse(w, "error authorizing", http.StatusUnauthorized, err)
		return
	}
	level.Debug(h.logger).Log("message", "reading response body")
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.errorResponse(w, "error reading authorization data", http.StatusInternalServerError, err)
		return
	}
	var cwr createWorkflowRequest
	err = json.Unmarshal(reqBody, &cwr)
	if err != nil {
		h.errorResponse(w, "error deserializing workflow data", http.StatusBadRequest, err)
		return
	}

	level.Debug(h.logger).Log("message", "validating workflow parameters")
	if err := h.validateWorkflowParameters(cwr.Parameters); err != nil {
		h.errorResponse(w, "error in parameters", http.StatusInternalServerError, err)
		return
	}

	frameworks, err := h.config.listFrameworks()
	if err != nil {
		h.errorResponse(w, "error reading frameworks from config", http.StatusBadRequest, err)
		return
	}
	if !stringInSlice(frameworks, cwr.Framework) {
		h.errorResponse(w, "unknown framework", http.StatusBadRequest, err)
		return
	}

	types, err := h.config.listTypes(cwr.Framework)
	if err != nil {
		h.errorResponse(w, "error reading types from config", http.StatusBadRequest, err)
		return
	}
	if !stringInSlice(types, cwr.Type) {
		h.errorResponse(w, "unknown type", http.StatusBadRequest, err)
		return
	}

	// TODO long term, we should evaluate if hard coding in code is the right approach to
	// specifying different argument types vs allowing dynmaic specification and
	// interpolation in service/config.yaml
	for k := range cwr.Arguments {
		if k != "init" && k != "execute" {
			h.errorResponse(w, "arguments must be init or execute", http.StatusBadRequest, err)
			return
		}
	}

	if !h.validateProjectName(cwr.ProjectName, w) {
		return
	}
	if !h.validateTargetName(cwr.TargetName, w) {
		return
	}

	// TODO: Fix type must be specified valication and add test
	//	h.errorResponse(w, "type must be specified", http.StatusBadRequest, err)
	//	return
	//}

	workflowFrom := fmt.Sprintf("workflowtemplate/%s", cwr.WorkflowTemplateName)
	executeContainerImageURI := cwr.Parameters["execute_container_image_uri"]
	environmentVariablesString := generateEnvVariablesString(cwr.EnvironmentVariables)

	level.Debug(h.logger).Log("message", "generating command to execute")
	commandDefinition, err := h.config.getCommandDefinition(cwr.Framework, cwr.Type)
	if err != nil {
		h.errorResponse(w, "unable to get command definition", http.StatusBadRequest, err)
		return
	}
	executeCommand, err := generateExecuteCommand(commandDefinition, environmentVariablesString, cwr.Arguments)
	if err != nil {
		h.errorResponse(w, "unable to generate command", http.StatusBadRequest, err)
		return
	}

	level.Debug(h.logger).Log("message", "creating new credentials provider")
	cp, err := h.newCredentialsProvider(*a)
	if err != nil {
		h.errorResponse(w, "error bad or unknown credentials provider", http.StatusInternalServerError, err)
		return
	}
	level.Debug(h.logger).Log("message", "getting credentials provider token")
	credentialsToken, err := cp.getToken()
	if err != nil {
		h.errorResponse(w, "error getting credentials provider token", http.StatusInternalServerError, err)
		return
	}

	projectExists, err := cp.projectExists(cwr.ProjectName)
	if err != nil {
		h.errorResponse(w, "error checking project", http.StatusInternalServerError, err)
		return
	}

	if !projectExists {
		h.errorResponse(w, "project does not exist", http.StatusBadRequest, err)
		return
	}

	// TODO: handle error when implemented
	//targetExists, _ := cp.targetExists(cwr.TargetName)
	//if !targetExists {
	//	h.errorResponse(w, "target must already exist", http.StatusBadRequest, err)
	//	return
	//}
	level.Debug(h.logger).Log("message", "creating workflow parameters")
	parameters := newWorkflowParameters(environmentVariablesString, executeCommand, executeContainerImageURI, cwr.TargetName, cwr.ProjectName, cwr.Parameters, credentialsToken)

	level.Debug(h.logger).Log("message", "creating workflow")
	workflowName, err := h.argo.Submit(workflowFrom, parameters)
	if err != nil {
		h.errorResponse(w, "error creating workflow", http.StatusInternalServerError, err)
		return
	}
	level.Debug(h.logger).Log("message", "workflow created", "workflow", workflowName)
	tokenHead := credentialsToken[0:8]
	level.Info(h.logger).Log("message", fmt.Sprintf("Received token '%s...'", tokenHead))
	var cwresp createWorkflowResponse
	cwresp.WorkflowName = workflowName
	jsonData, err := json.Marshal(cwresp)
	if err != nil {
		h.errorResponse(w, "error serializing workflow response", http.StatusBadRequest, err)
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
	level.Debug(h.logger).Log("message", "getting workflow status", "workflow", workflowName)
	status, err := h.argo.GetStatus(workflowName)
	if err != nil {
		h.errorResponse(w, "error getting workflow", http.StatusBadRequest, err)
		return
	}
	level.Debug(h.logger).Log("message", "decoding get workflow response", "workflow", workflowName)
	jsonData, err := json.Marshal(status) // TODO handle error in http resp
	if err != nil {
		h.errorResponse(w, "error serializing workflow", http.StatusBadRequest, err)
		return
	}
	fmt.Fprint(w, string(jsonData))
}

// Gets a workflow
func (h handler) getTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	targetName := vars["targetName"]
	ah := r.Header.Get("Authorization")

	level.Debug(h.logger).Log("message", "authorizing get target permissions")
	a, err := newAuthorization(ah) // todo add validation
	if err != nil {
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized, err)
		return
	}
	level.Debug(h.logger).Log("message", "validating authorized admin")
	if !a.authorizedAdmin() {
		h.errorResponse(w, "error must be authorized admin", http.StatusUnauthorized, nil)
		return
	}

	level.Debug(h.logger).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a)
	if err != nil {
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest, err)
		return
	}
	level.Debug(h.logger).Log("message", "getting target information", "project", projectName, "target", targetName)
	targetInfo, err := cp.getTarget(projectName, targetName)
	if err != nil {
		h.errorResponse(w, "error retrieving target information", http.StatusBadRequest, err)
		return
	}
	jsonResult, err := json.Marshal(targetInfo)
	if err != nil {
		h.errorResponse(w, "error retrieving json target data", http.StatusInternalServerError, err)
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
	level.Debug(h.logger).Log("message", "retrieving workflow logs", "workflow", workflowName)
	argoWorkflowLogs, err := h.argo.GetLogs(workflowName)
	if err != nil {
		h.errorResponse(w, "error getting workflow logs", http.StatusBadRequest, err)
		return
	}
	jsonData, err := json.Marshal(argoWorkflowLogs)
	if err != nil {
		h.errorResponse(w, "error serializing workflow logs", http.StatusInternalServerError, err)
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
	level.Debug(h.logger).Log("message", "retrieving workflow logs", "workflow", workflowName)
	err := h.argo.GetLogStream(workflowName, w)
	if err != nil {
		level.Error(h.logger).Log("message", "error getting workflow logstream", "error", err)
		h.errorResponse(w, "error getting workflow logs", http.StatusBadRequest, err)
		return
	}
}

// Returns a new token
func newArgoCloudOpsToken(provider, key, secret string) *token {
	return &token{
		Token: fmt.Sprintf("%s:%s:%s", provider, key, secret),
	}
}

func (h handler) requestLogger(r *http.Request, fields ...string) log.Logger {
	if r.Header.Get("X-TransactionID") != "" {
		return log.With(h.logger, "txid", r.Header.Get("X-TransactionID"))
	}
	return h.logger
}

// Creates a project
func (h handler) createProject(w http.ResponseWriter, r *http.Request) {
	l := h.requestLogger(r)
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.errorResponse(w, "error reading request body", http.StatusInternalServerError, nil)
		return
	}

	var capp createProjectRequest
	err = json.Unmarshal(reqBody, &capp)
	if err != nil {
		h.errorResponse(w, "error creating credentials provider", http.StatusInternalServerError, err)
		return
	}

	ah := r.Header.Get("Authorization")
	level.Debug(l).Log("message", "authorizing project creation")
	a, err := newAuthorization(ah) // todo add validation
	if err != nil {
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized, err)
		return
	}

	level.Debug(l).Log("message", "validating authorized admin")
	if !a.authorizedAdmin() {
		h.errorResponse(w, "error must be authorized admin", http.StatusUnauthorized, nil)
		return
	}

	level.Debug(l).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a)
	if err != nil {
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest, err)
		return
	}

	if !h.validateProjectName(capp.Name, w) {
		return
	}

	projectExists, err := cp.projectExists(capp.Name)
	if err != nil {
		h.errorResponse(w, "error checking project", http.StatusInternalServerError, err)
		return
	}

	if projectExists {
		h.errorResponse(w, "project already exists", http.StatusBadRequest, err)
		return
	}

	level.Debug(l).Log("message", "creating project")
	role, secret, err := cp.createProject(capp.Name)
	if err != nil {
		h.errorResponse(w, "error creating project", http.StatusInternalServerError, err)
		return
	}

	level.Debug(l).Log("message", "retrieving Argo CloudOps token")
	t := newArgoCloudOpsToken("vault", role, secret)
	jsonResult, err := json.Marshal(t)
	if err != nil {
		h.errorResponse(w, "error serializing token", http.StatusInternalServerError, err)
		return
	}
	fmt.Fprint(w, string(jsonResult))
}

// Get a project
func (h handler) getProject(w http.ResponseWriter, r *http.Request) {
	ah := r.Header.Get("Authorization")
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	level.Debug(h.logger).Log("message", "authorizing get project")
	a, err := newAuthorization(ah) // todo add validation
	if err != nil {
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized, err)
		return
	}
	level.Debug(h.logger).Log("message", "validating authorized admin")
	if !a.authorizedAdmin() {
		h.errorResponse(w, "error must be authorized admin", http.StatusUnauthorized, nil)
		return
	}

	level.Debug(h.logger).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a)
	if err != nil {
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest, err)
		return
	}

	level.Debug(h.logger).Log("message", "getting project", "project name", projectName)
	jsonResult, err := cp.getProject(projectName)
	if err != nil {
		h.errorResponse(w, "error retrieving project", http.StatusNotFound, err)
		return
	}
	fmt.Fprint(w, jsonResult)
}

// Delete a project
func (h handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	ah := r.Header.Get("Authorization")
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	level.Debug(h.logger).Log("message", "authorizing delete project")
	a, err := newAuthorization(ah) // todo add validation
	if err != nil {
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized, err)
		return
	}
	level.Debug(h.logger).Log("message", "validating authorized admin")
	if !a.authorizedAdmin() {
		h.errorResponse(w, "error must be authorized admin", http.StatusUnauthorized, nil)
		return
	}

	level.Debug(h.logger).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a)
	if err != nil {
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest, err)
		return
	}

	level.Debug(h.logger).Log("message", "checking if project exists", "project name", projectName)
	projectExists, err := cp.projectExists(projectName)
	if err != nil {
		h.errorResponse(w, "error checking project", http.StatusInternalServerError, err)
		return
	}

	if !projectExists {
		level.Debug(h.logger).Log("message", "no action required because project does not exist", "project name", projectName)
		return
	}

	level.Debug(h.logger).Log("message", "getting all targets in project", "project name", projectName)
	targets, err := cp.listTargets(projectName)
	if err != nil {
		h.errorResponse(w, "error getting all targets", http.StatusInternalServerError, err)
		return
	}

	if len(targets) > 0 {
		h.errorResponse(w, "project has existing targets, not deleting", http.StatusBadRequest, err)
		return
	}

	level.Debug(h.logger).Log("message", "deleting project", "project name", projectName)
	err = cp.deleteProject(projectName)
	if err != nil {
		h.errorResponse(w, "error deleting project", http.StatusBadRequest, err)
		return
	}
}

// Creates a target
func (h handler) createTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	ah := r.Header.Get("Authorization")
	level.Debug(h.logger).Log("message", "authorizing target creation")
	a, err := newAuthorization(ah)
	if err != nil {
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized, err)
		return
	}
	level.Debug(h.logger).Log("message", "validating authorized admin")
	if !a.authorizedAdmin() {
		h.errorResponse(w, "error must be authorized admin", http.StatusUnauthorized, err)
		return
	}
	level.Debug(h.logger).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a)
	if err != nil {
		h.errorResponse(w, "error creating credentials provider", http.StatusInternalServerError, err)
		return
	}
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.errorResponse(w, "error reading request body", http.StatusInternalServerError, err)
		return
	}
	var ctr createTargetRequest
	err = json.Unmarshal(reqBody, &ctr)
	if err != nil {
		h.errorResponse(w, "error serializing request body to target data", http.StatusBadRequest, err)
		return
	}
	level.Debug(h.logger).Log("message", "validating target name", "target name", ctr.Name)
	if !h.validateTargetName(ctr.Name, w) {
		return
	}
	if ctr.Type != "aws_account" {
		h.errorResponse(w, "type must be aws_account", http.StatusBadRequest, nil)
		return
	}
	if len(ctr.Properties.PolicyArns) > 5 {
		h.errorResponse(w, "policy arns list length cannot be greater than 5", http.StatusBadRequest, nil)
		return
	}
	for _, policyArn := range ctr.Properties.PolicyArns {
		if !arn.IsARN(policyArn) {
			h.errorResponse(w, "policy arn "+policyArn+" must be a valid arn", http.StatusBadRequest, nil)
			return
		}
	}
	if !arn.IsARN(ctr.Properties.RoleArn) {
		h.errorResponse(w, "role arn "+ctr.Properties.RoleArn+" must be a valid arn", http.StatusBadRequest, nil)
		return
	}
	targetExists, _ := cp.targetExists(ctr.Name)
	// TODO: handle error when implemented
	if targetExists {
		h.errorResponse(w, "target name must not already exist", http.StatusBadRequest, nil)
		return
	}
	level.Debug(h.logger).Log("message", "creating target", "project name", projectName, "target name", ctr.Name)
	err = cp.createTarget(projectName, ctr)
	if err != nil {
		h.errorResponse(w, "error creating target", http.StatusInternalServerError, err)
		return
	}
	fmt.Fprint(w, "{}")
}

// Deletes a target
func (h handler) deleteTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	targetName := vars["targetName"]
	ah := r.Header.Get("Authorization")

	level.Debug(h.logger).Log("message", "authorizing delete target permissions")
	a, err := newAuthorization(ah) // todo add validation
	if err != nil {
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized, err)
		return
	}
	level.Debug(h.logger).Log("message", "validating authorized admin")
	if !a.authorizedAdmin() {
		h.errorResponse(w, "error must be authorized admin", http.StatusUnauthorized, nil)
		return
	}

	level.Debug(h.logger).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a)
	if err != nil {
		h.errorResponse(w, "error creating credentials provider", http.StatusBadRequest, err)
		return
	}
	level.Debug(h.logger).Log("message", "deleting target", "project", projectName, "target", targetName)
	err = cp.deleteTarget(projectName, targetName)
	if err != nil {
		h.errorResponse(w, "error deleting target", http.StatusBadRequest, err)
		return
	}
}

// Lists the targets for a project
func (h handler) listTargets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	ah := r.Header.Get("Authorization")
	level.Debug(h.logger).Log("message", "authorizing target list retrieval")
	a, err := newAuthorization(ah)
	if err != nil {
		h.errorResponse(w, "error authorizing using Authorization header", http.StatusUnauthorized, err)
		return
	}
	level.Debug(h.logger).Log("message", "validating authorized admin")
	if !a.authorizedAdmin() {
		h.errorResponse(w, "error must be authorized admin", http.StatusUnauthorized, err)
		return
	}
	level.Debug(h.logger).Log("message", "creating credential provider")
	cp, err := h.newCredentialsProvider(*a)
	if err != nil {
		h.errorResponse(w, "error creating credentials provider", http.StatusInternalServerError, err)
		return
	}

	targets, err := cp.listTargets(projectName)
	if err != nil {
		h.errorResponse(w, "error listing targets", http.StatusInternalServerError, err)
		return
	}
	data, err := json.Marshal(targets)
	if err != nil {
		level.Error(h.logger).Log("message", "error deserializing targets", "error", err)
	}

	fmt.Fprint(w, string(data))
}

// Convenience method that writes a failure response in a standard manner
func (h handler) errorResponse(w http.ResponseWriter, message string, httpStatus int, err error) {
	level.Error(h.logger).Log("message", message, "error", err)
	r := generateErrorResponseJSON(message)
	w.WriteHeader(httpStatus)
	fmt.Fprint(w, r)
}

// Validates a project name
func (h handler) validateProjectName(projectName string, w http.ResponseWriter) bool {
	if len(projectName) < 4 {
		h.errorResponse(w, "project name must be greater than 3 characters", http.StatusBadRequest, nil)
		return false
	}
	if len(projectName) > 32 {
		h.errorResponse(w, "project name must be less than 32 characters", http.StatusBadRequest, nil)
		return false
	}
	if !isStringAlphaNumeric(projectName) {
		h.errorResponse(w, "project name must be alpha-numeric", http.StatusBadRequest, nil)
		return false
	}
	return true
}

// Validates a target name
func (h handler) validateTargetName(targetName string, w http.ResponseWriter) bool {
	if len(targetName) < 4 {
		h.errorResponse(w, "target name must be greater than 3 characters", http.StatusBadRequest, nil)
		return false
	}
	if len(targetName) > 32 {
		h.errorResponse(w, "target name must be less than 32 characters", http.StatusBadRequest, nil)
		return false
	}
	if !isStringAlphaNumericUnderscore(targetName) {
		h.errorResponse(w, "target name must be alpha-numeric with underscores", http.StatusBadRequest, nil)
		return false
	}
	return true
}

// TODO: Fix to include -
// Validates a workflow name
//func (h handler) validateWorkflowName(workflowName string, w http.ResponseWriter) bool {
//	return h.validateName(workflowName, "workflow name", w)
//}

// Validates name according to naming rules:
// 1. Must be alphanumeric
// 2. Must have a minimum length of 4
// 3. Must have a maximum length of 32
func (h handler) validateName(name string, desc string, w http.ResponseWriter) bool {
	level.Debug(h.logger).Log("message", "validating "+desc, desc, name)
	if !isStringAlphaNumericUnderscore(name) {
		h.errorResponse(w, desc+" must be alpha-numeric", http.StatusBadRequest, nil)
		return false
	}
	if len(name) < 4 {
		h.errorResponse(w, desc+" must be greater than 3 characters", http.StatusBadRequest, nil)
		return false
	}
	if len(name) > 32 {
		h.errorResponse(w, desc+" must be less than 32 characters", http.StatusBadRequest, nil)
		return false
	}
	return true
}

// Returns true, if the image uri is a valid container image uri
func (h handler) isValidImageUri(imageUri string) bool {
	_, err := reference.ParseAnyReference(imageUri)
	if err != nil {
		return false
	}
	return true
}
