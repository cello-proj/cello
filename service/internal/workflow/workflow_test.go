package workflow

import (
	"context"
	"errors"
	"fmt"
	"testing"

	argoWorkflowAPIClient "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	mockArgoWorkflowAPIClient "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow/mocks"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestArgoWorkflowsListStatus(t *testing.T) {
	tests := []struct {
		name             string
		workflowList     *v1alpha1.WorkflowList
		listWorkflowsErr error
		expectedStatus   []Status
		errExpected      bool
	}{
		{
			name: "list workflows success",
			workflowList: &v1alpha1.WorkflowList{
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
			workflowList:     nil,
			listWorkflowsErr: errors.New("list workflows error"),
			expectedStatus:   []Status{},
			errExpected:      true,
		},
		{
			name:             "list status empty",
			workflowList:     new(v1alpha1.WorkflowList),
			listWorkflowsErr: nil,
			expectedStatus:   []Status{},
			errExpected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockArgoWorkflowAPIClient.WorkflowServiceClient{}
			mockClient.On("ListWorkflows", mock.Anything, &argoWorkflowAPIClient.WorkflowListRequest{Namespace: "namespace"}).
				Return(tt.workflowList, tt.listWorkflowsErr)

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
		name           string
		workflowName   string
		workflow       *v1alpha1.Workflow
		getWorkflowErr error
		expectedStatus *Status
		errExpected    bool
	}{
		{
			name:         "get status",
			workflowName: "testWorkflow1",
			workflow: &v1alpha1.Workflow{
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
			name:           "get status error",
			workflowName:   "testWorkflow1",
			workflow:       new(v1alpha1.Workflow),
			getWorkflowErr: errors.New("get workflow error"),
			expectedStatus: nil,
			errExpected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockArgoWorkflowAPIClient.WorkflowServiceClient{}
			mockClient.On("GetWorkflow", mock.Anything, &argoWorkflowAPIClient.WorkflowGetRequest{Name: tt.workflowName, Namespace: "namespace"}).
				Return(tt.workflow, tt.getWorkflowErr)

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
		name      string
		err       error
		result    string
		errResult error
	}{
		{
			name:   "submit workflow",
			result: "testworkflow1",
		},
		{
			name:      "get workflow logs error",
			err:       fmt.Errorf("submit error"),
			errResult: fmt.Errorf("failed to submit workflow: submit error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argoWf := NewArgoWorkflow(
				mockArgoClient{err: tt.err},
				"namespace",
			)

			workflow, err := argoWf.Submit(context.Background(), "test/test", map[string]string{"param": "value"}, map[string]string{"X-B3-TraceId": "test-txid"})
			if err != nil {
				if tt.errResult != nil && tt.errResult.Error() != err.Error() {
					t.Errorf("\nwant: %v\n got: %v", tt.errResult, err)
				}
				if tt.errResult == nil {
					t.Errorf("\nwant: %v\n got: %v", tt.errResult, err)
				}
			} else {
				if !cmp.Equal(workflow, tt.result) {
					t.Errorf("\nwant: %v\n got: %v", tt.result, workflow)
				}
			}
		})
	}
}

type mockArgoClient struct {
	argoWorkflowAPIClient.WorkflowServiceClient
	status v1alpha1.WorkflowPhase
	err    error
}

func (m mockArgoClient) SubmitWorkflow(ctx context.Context, in *argoWorkflowAPIClient.WorkflowSubmitRequest, opts ...grpc.CallOption) (*v1alpha1.Workflow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &v1alpha1.Workflow{TypeMeta: v1.TypeMeta{}, ObjectMeta: v1.ObjectMeta{Name: "testWorkflow1"}, Status: v1alpha1.WorkflowStatus{Phase: m.status}}, nil
}
