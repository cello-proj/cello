package workflow

import (
	"context"
	"fmt"
	"testing"

	argoWorkflowAPIClient "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestArgoWorkflowsList(t *testing.T) {
	tests := []struct {
		name      string
		output    []string
		listErr   error
		errResult error
	}{
		{
			name:      "list workflows success",
			output:    []string{"testWorkflow1"},
			errResult: nil,
		},
		{
			name:      "list workflows error",
			output:    []string{},
			listErr:   fmt.Errorf("list error"),
			errResult: fmt.Errorf("list error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argoWf := NewArgoWorkflow(
				mockArgoClient{err: tt.listErr},
				"namespace",
			)

			out, err := argoWf.List(context.Background())

			if err != nil {
				if tt.errResult != nil && tt.errResult.Error() != err.Error() {
					t.Errorf("\nwant: %v\n got: %v", tt.errResult, err)
				}

				if tt.errResult == nil {
					t.Errorf("\nwant: %v\n got: %v", tt.errResult, err)
				}
			}

			if err == nil && tt.errResult != nil {
				t.Errorf("\nwant: %v\n got: %v", tt.errResult, err)
			}

			if !cmp.Equal(out, tt.output) {
				t.Errorf("\nwant: %v\n got: %v", tt.output, out)
			}
		})
	}
}

func TestArgoStatus(t *testing.T) {
	tests := []struct {
		name               string
		argoWorkflowStatus v1alpha1.WorkflowPhase
		statusErr          error
		result             string
		errResult          error
	}{
		{
			name:               "get workflow status pending",
			argoWorkflowStatus: v1alpha1.WorkflowPending,
			result:             "pending",
		},
		{
			name:               "get workflow status running",
			argoWorkflowStatus: v1alpha1.WorkflowRunning,
			result:             "running",
		},
		{
			name:               "get workflow status failed",
			argoWorkflowStatus: v1alpha1.WorkflowFailed,
			result:             "failed",
		},
		{
			name:               "get workflow status succeeded",
			argoWorkflowStatus: v1alpha1.WorkflowSucceeded,
			result:             "succeeded",
		},
		{
			name:      "get workflow status error",
			statusErr: fmt.Errorf("status error"),
			errResult: fmt.Errorf("status error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argoWf := NewArgoWorkflow(
				mockArgoClient{status: tt.argoWorkflowStatus, err: tt.statusErr},
				"namespace",
			)

			status, err := argoWf.Status(context.Background(), "workflow")
			if err != nil {
				if tt.errResult != nil && tt.errResult.Error() != err.Error() {
					t.Errorf("\nwant: %v\n got: %v", tt.errResult, err)
				}

				if tt.errResult == nil {
					t.Errorf("\nwant: %v\n got: %v", tt.errResult, err)
				}
			} else {
				if !cmp.Equal(status.Status, tt.result) {
					t.Errorf("\nwant: %v\n got: %v", tt.result, status.Status)
				}
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

func (m mockArgoClient) ListWorkflows(ctx context.Context, in *argoWorkflowAPIClient.WorkflowListRequest, opts ...grpc.CallOption) (*v1alpha1.WorkflowList, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &v1alpha1.WorkflowList{Items: []v1alpha1.Workflow{
		{TypeMeta: v1.TypeMeta{}, ObjectMeta: v1.ObjectMeta{Name: "testWorkflow1"}}}}, nil
}

func (m mockArgoClient) GetWorkflow(ctx context.Context, in *argoWorkflowAPIClient.WorkflowGetRequest, opts ...grpc.CallOption) (*v1alpha1.Workflow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &v1alpha1.Workflow{TypeMeta: v1.TypeMeta{}, ObjectMeta: v1.ObjectMeta{Name: "testWorkflow1"}, Status: v1alpha1.WorkflowStatus{Phase: m.status}}, nil
}

func (m mockArgoClient) SubmitWorkflow(ctx context.Context, in *argoWorkflowAPIClient.WorkflowSubmitRequest, opts ...grpc.CallOption) (*v1alpha1.Workflow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &v1alpha1.Workflow{TypeMeta: v1.TypeMeta{}, ObjectMeta: v1.ObjectMeta{Name: "testWorkflow1"}, Status: v1alpha1.WorkflowStatus{Phase: m.status}}, nil
}
