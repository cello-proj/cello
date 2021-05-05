package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	argoWorkflowAPIClient "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	argoWorkflowAPISpec "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

const mainContainer = "main"

// Workflow interface is used for interacting with workflow services.
type Workflow interface {
	ListWorkflows() ([]string, error)
	GetStatus(workflowName string) (*workflowStatus, error)
	GetLogs(workflowName string) (*workflowLogs, error)
	GetLogStream(workflowName string, data http.ResponseWriter) error
	Submit(from string, parameters map[string]string) (string, error)
}

func newArgoWorkflow() Workflow {
	argoCtx, argoClient := client.NewAPIClient()
	return &argoWorkflow{
		argoCtx,
		argoClient.NewWorkflowServiceClient(),
	}
}

type argoWorkflow struct {
	argoCtx context.Context
	argoSvc argoWorkflowAPIClient.WorkflowServiceClient
}

// workflowLogs TODO doc
type workflowLogs struct {
	Logs []string `json:"logs"`
}

func (a argoWorkflow) ListWorkflows() ([]string, error) {
	workflowIDs := []string{}

	workflowListResult, err := a.argoSvc.ListWorkflows(a.argoCtx, &argoWorkflowAPIClient.WorkflowListRequest{
		Namespace: argoNamespace(),
	})

	if err != nil {
		return workflowIDs, err
	}

	for _, item := range workflowListResult.Items {
		workflowIDs = append(workflowIDs, item.ObjectMeta.Name)
	}

	return workflowIDs, nil
}

type workflowStatus struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Finished string `json:"finished"`
}

func (a argoWorkflow) GetStatus(workflowName string) (*workflowStatus, error) {
	workflow, err := a.argoSvc.GetWorkflow(a.argoCtx, &argoWorkflowAPIClient.WorkflowGetRequest{
		Name:      workflowName,
		Namespace: argoNamespace(),
	})
	if err != nil {
		return nil, err
	}
	workflowData := workflowStatus{
		Name:     workflowName,
		Status:   strings.ToLower(string(workflow.Status.Phase)),
		Created:  fmt.Sprint(workflow.CreationTimestamp.Unix()),
		Finished: fmt.Sprint(workflow.Status.FinishedAt.Unix()),
	}

	return &workflowData, nil
}

func (a argoWorkflow) GetLogs(workflowName string) (*workflowLogs, error) {
	stream, err := a.argoSvc.WorkflowLogs(a.argoCtx, &argoWorkflowAPIClient.WorkflowLogRequest{
		Name:      workflowName,
		Namespace: argoNamespace(),
		LogOptions: &v1.PodLogOptions{
			Container: mainContainer,
		},
	})
	if err != nil {
		return nil, err
	}

	var argoWorkflowLogs workflowLogs
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		argoWorkflowLogs.Logs = append(argoWorkflowLogs.Logs, fmt.Sprintf("%s: %s", event.PodName, event.Content))
	}

	return &argoWorkflowLogs, nil
}

func (a argoWorkflow) GetLogStream(workflowName string, w http.ResponseWriter) error {
	stream, err := a.argoSvc.WorkflowLogs(a.argoCtx, &argoWorkflowAPIClient.WorkflowLogRequest{
		Name:      workflowName,
		Namespace: argoNamespace(),
		LogOptions: &v1.PodLogOptions{
			Container: mainContainer,
			Follow:    true,
		},
	})
	if err != nil {
		return err
	}

	clientGone := w.(http.CloseNotifier).CloseNotify()
	for {
		select {
		case <-clientGone:
			return nil
		default:
			event, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			fmt.Fprintf(w, fmt.Sprintf("%s: %s\n", event.GetPodName(), event.GetContent()))
			w.(http.Flusher).Flush()
			status, err := a.GetStatus(workflowName)
			if err != nil {
				return err
			}
			if event == nil && status.Status != "running" && status.Status != "pending" {
				return nil
			}
		}
	}
}

func (a argoWorkflow) Submit(from string, parameters map[string]string) (string, error) {
	parts := strings.SplitN(from, "/", 2)
	for _, part := range parts {
		if part == "" {
			return "", fmt.Errorf("resource identifier '%s' is malformed. Should be `kind/name`, e.g. cronwf/hello-world-cwf", from)
		}
	}
	kind := parts[0]
	name := parts[1]

	var parameterStrings []string
	for k, v := range parameters {
		parameterStrings = append(parameterStrings, fmt.Sprintf("%s=%s", k, v))
	}

	generateNamePrefix := fmt.Sprintf("%s-%s-", parameters["project_name"], parameters["target_name"])

	created, err := a.argoSvc.SubmitWorkflow(a.argoCtx, &argoWorkflowAPIClient.WorkflowSubmitRequest{
		Namespace:    argoNamespace(),
		ResourceKind: kind,
		ResourceName: name,
		SubmitOptions: &argoWorkflowAPISpec.SubmitOpts{
			GenerateName: generateNamePrefix,
			Parameters:   parameterStrings,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to submit workflow: %v", err)
	}

	return strings.ToLower(created.Name), nil
}

func newWorkflowParameters(environmentVariablesString, executeCommand, executeContainerImageURI, targetName, projectName string, cliParameters map[string]string, credentialsToken string) map[string]string {
	parameters := map[string]string{
		"environment_variables_string": environmentVariablesString,
		"execute_command":              executeCommand,
		"execute_container_image_uri":  executeContainerImageURI,
		"project_name":                 projectName,
		"target_name":                  targetName,
		"credentials_token":            credentialsToken,
	}

	// this include override parameters
	// don't want to necessarily allow overriding everything
	// for now, constrainting to execute image uri
	// TODO find a dynamic way to combine two json objects
	// Either do it here or after it is generated and passed to argoWorkflow submit
	for k, v := range cliParameters {
		if k == "execute_container_image_uri" {
			parameters["execute_container_image_uri"] = v
		}

		if k == "pre_container_image_uri" {
			parameters["pre_container_image_uri"] = v
		}
	}

	return parameters
}

type createWorkflowResponse struct {
	WorkflowName string `json:"workflow_name"`
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
