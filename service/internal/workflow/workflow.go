package workflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	argoWorkflowAPIClient "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	argoWorkflowAPISpec "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const mainContainer = "main"

// Workflow interface is used for interacting with workflow services.
type Workflow interface {
	List(ctx context.Context) ([]string, error)
	Logs(ctx context.Context, workflowName string) (*Logs, error)
	LogStream(ctx context.Context, workflowName string, data http.ResponseWriter) error
	Status(ctx context.Context, workflowName string) (*Status, error)
	Submit(ctx context.Context, from string, parameters map[string]string, labels map[string]string) (string, error)
}

// NewArgoWorkflow creates an Argo workflow.
func NewArgoWorkflow(cl argoWorkflowAPIClient.WorkflowServiceClient, n string) Workflow {
	return &ArgoWorkflow{
		namespace: n,
		svc:       cl,
	}
}

// ArgoWorkflow represents an Argo Workflow.
type ArgoWorkflow struct {
	namespace string
	svc       argoWorkflowAPIClient.WorkflowServiceClient
}

// Logs represents workflow logs.
type Logs struct {
	Logs []string `json:"logs"`
}

// List returns a list of workflows.
func (a ArgoWorkflow) List(ctx context.Context) ([]string, error) {
	workflowIDs := []string{}

	workflowListResult, err := a.svc.ListWorkflows(ctx, &argoWorkflowAPIClient.WorkflowListRequest{
		Namespace: a.namespace,
	})

	if err != nil {
		return workflowIDs, err
	}

	for _, item := range workflowListResult.Items {
		workflowIDs = append(workflowIDs, item.ObjectMeta.Name)
	}

	return workflowIDs, nil
}

// Status represents a workflow status.
type Status struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Finished string `json:"finished"`
}

// Status returns a workflow status.
func (a ArgoWorkflow) Status(ctx context.Context, workflowName string) (*Status, error) {
	workflow, err := a.svc.GetWorkflow(ctx, &argoWorkflowAPIClient.WorkflowGetRequest{
		Name:      workflowName,
		Namespace: a.namespace,
	})

	if err != nil {
		return nil, err
	}

	workflowData := Status{
		Name:     workflowName,
		Status:   strings.ToLower(string(workflow.Status.Phase)),
		Created:  fmt.Sprint(workflow.CreationTimestamp.Unix()),
		Finished: fmt.Sprint(workflow.Status.FinishedAt.Unix()),
	}

	return &workflowData, nil
}

// Logs returns logs for a workflow.
func (a ArgoWorkflow) Logs(ctx context.Context, workflowName string) (*Logs, error) {
	stream, err := a.svc.WorkflowLogs(ctx, &argoWorkflowAPIClient.WorkflowLogRequest{
		Name:      workflowName,
		Namespace: a.namespace,
		LogOptions: &v1.PodLogOptions{
			Container: mainContainer,
		},
	})

	if err != nil {
		return nil, err
	}

	var argoWorkflowLogs Logs
	for {
		event, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		argoWorkflowLogs.Logs = append(argoWorkflowLogs.Logs, fmt.Sprintf("%s: %s", event.PodName, event.Content))
	}

	return &argoWorkflowLogs, nil
}

// LogStream returns a log stream for a workflow.
func (a ArgoWorkflow) LogStream(ctx context.Context, workflowName string, w http.ResponseWriter) error {
	stream, err := a.svc.WorkflowLogs(ctx, &argoWorkflowAPIClient.WorkflowLogRequest{
		Name:      workflowName,
		Namespace: a.namespace,
		LogOptions: &v1.PodLogOptions{
			Container: mainContainer,
			Follow:    true,
		},
	})

	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			event, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}

			if err != nil {
				return err
			}

			fmt.Fprintf(w, "%s: %s\n", event.GetPodName(), event.GetContent())
			w.(http.Flusher).Flush()
			status, err := a.Status(ctx, workflowName)
			if err != nil {
				return err
			}

			if event == nil && status.Status != "running" && status.Status != "pending" {
				return nil
			}
		}
	}
}

// Submit submits a workflow execution.
func (a ArgoWorkflow) Submit(ctx context.Context, from string, parameters map[string]string, workflowLabels map[string]string) (string, error) {
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

	created, err := a.svc.SubmitWorkflow(ctx, &argoWorkflowAPIClient.WorkflowSubmitRequest{
		Namespace:    a.namespace,
		ResourceKind: kind,
		ResourceName: name,
		SubmitOptions: &argoWorkflowAPISpec.SubmitOpts{
			GenerateName: generateNamePrefix,
			Parameters:   parameterStrings,
			Labels:       labels.FormatLabels(workflowLabels),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to submit workflow: %w", err)
	}

	return strings.ToLower(created.Name), nil
}

// NewParameters creates workflow parameters.
func NewParameters(environmentVariablesString, executeCommand, executeContainerImageURI, targetName, projectName string, cliParameters map[string]string, credentialsToken string) map[string]string {
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

// CreateWorkflowResponse creates a workflow response.
type CreateWorkflowResponse struct {
	WorkflowName string `json:"workflow_name"`
}
