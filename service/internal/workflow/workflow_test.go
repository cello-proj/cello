package workflow

import (
	"context"
	"errors"
	"testing"

	mockArgoWorkflowAPIClient "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow/mocks"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestArgoWorkflowsListStatus(t *testing.T) {
	tests := []struct {
		name             string
		workflowListResp *v1alpha1.WorkflowList
		listWorkflowsErr error
		expectedStatus   []Status
		errExpected      bool
	}{
		{
			name: "list workflows success",
			workflowListResp: &v1alpha1.WorkflowList{
				Items: []v1alpha1.Workflow{
					{
						ObjectMeta: v1.ObjectMeta{
							Name:              "testWorkflow1",
							CreationTimestamp: v1.Unix(1658514000, 0),
						},
						Status: v1alpha1.WorkflowStatus{
							Phase: v1alpha1.WorkflowRunning,
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name:              "testWorkflow2",
							CreationTimestamp: v1.Unix(1658512485, 0),
						},
						Status: v1alpha1.WorkflowStatus{
							Phase:      v1alpha1.WorkflowSucceeded,
							FinishedAt: v1.Unix(1658512623, 0),
						},
					},
				},
			},
			listWorkflowsErr: nil,
			expectedStatus: []Status{
				{
					Name:    "testWorkflow1",
					Status:  "running",
					Created: "1658514000",
				},
				{
					Name:     "testWorkflow2",
					Status:   "succeeded",
					Created:  "1658512485",
					Finished: "1658512623",
				},
			},
			errExpected: false,
		},
		{
			name:             "list status error",
			workflowListResp: nil,
			listWorkflowsErr: errors.New("list workflows error"),
			expectedStatus:   []Status{},
			errExpected:      true,
		},
		{
			name:             "list status empty",
			workflowListResp: new(v1alpha1.WorkflowList),
			listWorkflowsErr: nil,
			expectedStatus:   []Status{},
			errExpected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockArgoWorkflowAPIClient.WorkflowServiceClient{}
			mockClient.On("ListWorkflows", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*workflow.WorkflowListRequest")).
				Return(tt.workflowListResp, tt.listWorkflowsErr)

			argoWf := NewArgoWorkflow(
				mockClient,
				"namespace",
			)

			out, err := argoWf.ListStatus(context.Background())
			if err != nil {
				if !tt.errExpected {
					t.Errorf("\nerror not expected: %v", err)
				}
			}

			if !cmp.Equal(out, tt.expectedStatus) {
				t.Errorf("\nwant: %v\n got: %v", tt.expectedStatus, out)
			}
		})
	}
}

func TestArgoStatus(t *testing.T) {
	tests := []struct {
		name            string
		workflowName    string
		getWorkflowResp *v1alpha1.Workflow
		getWorkflowErr  error
		expectedStatus  *Status
		errExpected     bool
	}{
		{
			name:         "get status",
			workflowName: "testWorkflow1",
			getWorkflowResp: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:              "testWorkflow1",
					CreationTimestamp: v1.Unix(1658514000, 0),
				},
				Status: v1alpha1.WorkflowStatus{
					Phase:      v1alpha1.WorkflowSucceeded,
					FinishedAt: v1.Unix(1658512623, 0),
				},
			},
			getWorkflowErr: nil,
			expectedStatus: &Status{
				Name:     "testWorkflow1",
				Status:   "succeeded",
				Created:  "1658514000",
				Finished: "1658512623",
			},
			errExpected: false,
		},
		{
			name:            "get status error",
			workflowName:    "testWorkflow1",
			getWorkflowResp: new(v1alpha1.Workflow),
			getWorkflowErr:  errors.New("get workflow error"),
			expectedStatus:  nil,
			errExpected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockArgoWorkflowAPIClient.WorkflowServiceClient{}
			mockClient.On("GetWorkflow", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*workflow.WorkflowGetRequest")).
				Return(tt.getWorkflowResp, tt.getWorkflowErr)

			argoWf := NewArgoWorkflow(
				mockClient,
				"namespace",
			)

			status, err := argoWf.Status(context.Background(), "testWorkflow1")
			if err != nil {
				if !tt.errExpected {
					t.Errorf("\nerror not expected: %v", err)
				}
			}

			if !cmp.Equal(status, tt.expectedStatus) {
				t.Errorf("\nwant: %v\n got: %v", tt.expectedStatus, status)
			}
		})
	}
}

func TestArgoSubmit(t *testing.T) {
	tests := []struct {
		name               string
		submitWorkflowResp *v1alpha1.Workflow
		submitWorkflowErr  error
		expected           string
		errExpected        bool
	}{
		{
			name: "submit workflow",
			submitWorkflowResp: &v1alpha1.Workflow{
				ObjectMeta: v1.ObjectMeta{
					Name:              "testWorkflow1",
					CreationTimestamp: v1.Unix(1658514000, 0),
				},
				Status: v1alpha1.WorkflowStatus{
					Phase:      v1alpha1.WorkflowSucceeded,
					FinishedAt: v1.Unix(1658512623, 0),
				},
			},
			expected:    "testworkflow1",
			errExpected: false,
		},
		{
			name:               "submit workflow error",
			submitWorkflowResp: new(v1alpha1.Workflow),
			submitWorkflowErr:  errors.New("submit workflow error"),
			expected:           "",
			errExpected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockArgoWorkflowAPIClient.WorkflowServiceClient{}
			mockClient.On("SubmitWorkflow", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*workflow.WorkflowSubmitRequest")).
				Return(tt.submitWorkflowResp, tt.submitWorkflowErr)

			argoWf := NewArgoWorkflow(
				mockClient,
				"namespace",
			)

			workflowName, err := argoWf.Submit(context.Background(), "test/test", map[string]string{"param": "value"}, map[string]string{"X-B3-TraceId": "test-txid"})
			if err != nil {
				if !tt.errExpected {
					t.Errorf("\nerror not expected: %v", err)
				}
			}

			if !cmp.Equal(workflowName, tt.expected) {
				t.Errorf("\nwant: %v\n got: %v", tt.expected, workflowName)
			}
		})
	}
}

func TestNewParameters(t *testing.T) {
	environmentVariablesString := "ENVIRONMENT: prd"
	executeCommand := "fake_execution_command"
	executeContainerImageURI := "fake_execute_container_image_url"
	targetName := "fake_target_name"
	projectName := "fake_project_name"
	preContainerImageURI := "fake_pre_container_image_uri"
	credentialsToken := "fake_token"
	flowType := "sync"

	tests := []struct {
		name                       string
		environmentVariablesString string
		executeCommand             string
		executeContainerImageURI   string
		targetName                 string
		projectName                string
		cliParameters              map[string]string
		credentialsToken           string
		flowType                   string
		expected                   map[string]string
	}{
		{
			name:                       "new parameter",
			environmentVariablesString: environmentVariablesString,
			executeCommand:             executeCommand,
			executeContainerImageURI:   executeContainerImageURI,
			targetName:                 targetName,
			projectName:                projectName,
			cliParameters:              map[string]string{"pre_container_image_uri": preContainerImageURI},
			credentialsToken:           credentialsToken,
			flowType:                   flowType,
			expected: map[string]string{
				"environment_variables_string": environmentVariablesString,
				"execute_command":              executeCommand,
				"execute_container_image_uri":  executeContainerImageURI,
				"project_name":                 projectName,
				"target_name":                  targetName,
				"credentials_token":            credentialsToken,
				"type":                         flowType,
				"pre_container_image_uri":      preContainerImageURI,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parameters := NewParameters(
				tt.environmentVariablesString,
				tt.executeCommand,
				tt.executeContainerImageURI,
				tt.targetName,
				tt.projectName,
				tt.cliParameters,
				tt.credentialsToken,
				tt.flowType,
			)

			if !cmp.Equal(parameters, tt.expected) {
				t.Errorf("\nwant: %v\n got: %v", tt.expected, parameters)
			}
		})
	}
}
