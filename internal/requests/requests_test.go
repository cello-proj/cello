package requests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWorkflowValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateWorkflow
		types   []string
		wantErr error
	}{
		{
			name: "valid minimal",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
		},
		{
			name: "valid full",
			req: CreateWorkflow{
				Arguments: map[string][]string{
					"execute": {"--foo", "--bar"},
					"init":    {"--baz", "blah"},
				},
				EnvironmentVariables: map[string]string{
					"FOO": "BAR",
				},
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
					"pre_container_image_uri":     "argoproj-labs/argo-cloudops-pre",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
		},
		{
			name: "valid parameters no pre container image",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
		},
		{
			name: "too many arguments",
			req: CreateWorkflow{
				Arguments: map[string][]string{
					"execute": {"--foo", "--bar"},
					"init":    {"--baz", "blah"},
					"other":   {"--not", "valid"},
				},
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("arguments must be one of 'execute init'"),
		},
		{
			name: "not execute or init arguments",
			req: CreateWorkflow{
				Arguments: map[string][]string{
					"execute": {"--foo", "--bar"},
					"other":   {"--not", "valid"},
				},
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("arguments must be one of 'execute init'"),
		},
		{
			name: "missing parameters",
			req: CreateWorkflow{
				Framework:            "cdk",
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("parameter execute_container_image_uri is required"),
		},
		{
			name: "parameters missing exec container",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"pre_container_image_uri": "argoproj-labs/argo-cloudops-pre",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("parameter execute_container_image_uri is required"),
		},
		{
			name: "parameters bad exec container format",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "./foo/bar",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("parameter execute_container_image_uri must be a valid container uri"),
		},
		{
			name: "parameters bad pre container format",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
					"pre_container_image_uri":     "./foo/bar",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("parameter pre_container_image_uri must be a valid container uri"),
		},
		{
			name: "missing framework",
			req: CreateWorkflow{
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("framework is required"),
		},
		{
			name: "missing type",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("type is required"),
		},
		{
			name: "missing project name",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("project_name is required"),
		},
		{
			name: "too short project name",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "abc",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("project_name must be between 4 and 32 characters"),
		},
		{
			name: "too long project name",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "a12345678901234567890123456789012",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("project_name must be between 4 and 32 characters"),
		},
		{
			name: "invalid chars in project name",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "this-is-invalid",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("project_name must be alphanumeric"),
		},
		{
			name: "missing target name",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("target_name is required"),
		},
		{
			name: "too short target name",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "abc",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("target_name must be between 4 and 32 characters"),
		},
		{
			name: "too long target name",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "a12345678901234567890123456789012",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("target_name must be between 4 and 32 characters"),
		},
		{
			name: "invalid chars in target name",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1this",
				TargetName:           "this-is-invalid",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("target_name must be alphanumeric underscore"),
		},
		{
			name: "missing workflow template name",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
					"pre_container_image_uri":     "./foo/bar",
				},
				ProjectName: "project1",
				TargetName:  "target1",
				Type:        "diff",
			},
			wantErr: errors.New("workflow_template_name is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr != nil {
				assert.EqualError(t, tt.req.Validate(), tt.wantErr.Error())
			} else {
				assert.Equal(t, tt.wantErr, tt.req.Validate())
			}
		})
	}
}

func TestCreateWorkflowValidateType(t *testing.T) {
	tests := []struct {
		name    string
		types   []string
		wantErr error
	}{
		{
			name:  "valid",
			types: []string{"foo", "bar"},
		},
		{
			name:    "invalid",
			types:   []string{"bad"},
			wantErr: errors.New("type must be one of 'bad'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateWorkflow{
				Type: "foo",
			}
			assert.Equal(t, tt.wantErr, req.ValidateType(tt.types)())
		})
	}
}

func TestCreateTargetValidateTargetProperties(t *testing.T) {
	tests := []struct {
		name       string
		properites TargetProperties
		wantErr    error
	}{
		{
			name: "valid",
			properites: TargetProperties{
				CredentialType: "vault",
				RoleArn:        "arn:aws:iam::012345678901:role/test-role",
			},
		},
		{
			name: "invalid_credential_type",
			properites: TargetProperties{
				CredentialType: "boom",
				RoleArn:        "arn:aws:iam::012345678901:role/test-role",
			},
			wantErr: errors.New("credential_type must be one of 'vault'"),
		},
		{
			name: "invalid_role_arn",
			properites: TargetProperties{
				CredentialType: "vault",
				RoleArn:        "boom",
			},
			wantErr: errors.New("role_arn must be a valid arn"),
		},
		{
			name: "invalid_role_policies_arn",
			properites: TargetProperties{
				CredentialType: "vault",
				RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				PolicyArns: []string{
					"boom",
				},
			},
			wantErr: errors.New("policy_arns contains an invalid arn"),
		},
		{
			name: "invalid_too_many_role_policies",
			properites: TargetProperties{
				CredentialType: "vault",
				RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				PolicyArns: []string{
					"arn:aws:iam::012345678901:role/test-role",
					"arn:aws:iam::012345678901:role/test-role",
					"arn:aws:iam::012345678901:role/test-role",
					"arn:aws:iam::012345678901:role/test-role",
					"arn:aws:iam::012345678901:role/test-role",
					"arn:aws:iam::012345678901:role/test-role",
				},
			},
			wantErr: errors.New("policy_arns cannot be more than 5"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTarget{
				Properties: tt.properites,
			}
			assert.Equal(t, tt.wantErr, req.validateTargetProperties())
		})
	}
}
