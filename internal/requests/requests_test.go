package requests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/argoproj-labs/argo-cloudops/internal/validations"
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
			name: "only execute argument",
			req: CreateWorkflow{
				Arguments: map[string][]string{
					"execute": {"--foo", "--bar"},
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
		},
		{
			name: "only init argument",
			req: CreateWorkflow{
				Arguments: map[string][]string{
					"init": {"--foo", "--bar"},
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
			name: "rejects unapproved exec container",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "badactor-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("parameter execute_container_image_uri must be an approved image uri"),
		},
		{
			name: "rejects unapproved pre container",
			req: CreateWorkflow{
				Framework: "cdk",
				Parameters: map[string]string{
					"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
					"pre_container_image_uri":     "badactor-labs/argo-cloudops-exec",
				},
				ProjectName:          "project1",
				TargetName:           "target1",
				Type:                 "diff",
				WorkflowTemplateName: "template1",
			},
			wantErr: errors.New("parameter pre_container_image_uri must be an approved image uri"),
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

	validations.SetImageURIs([]string{"argoproj-labs/*"})

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

func TestCreateGitWorkflowValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateGitWorkflow
		types   []string
		wantErr error
	}{
		{
			name: "valid",
			req: CreateGitWorkflow{
				CommitHash: "8458fd753f9fde51882414564c20df6d4c34a90e",
				Path:       "./manifest.yaml",
			},
		},
		{
			name: "missing commit hash",
			req: CreateGitWorkflow{
				Path: "./manifest.yaml",
			},
			wantErr: errors.New("sha is required"),
		},
		{
			name: "commit hash must be alphanumeric",
			req: CreateGitWorkflow{
				CommitHash: "8--",
				Path:       "./manifest.yaml",
			},
			wantErr: errors.New("sha must be alphanumeric"),
		},
		{
			name: "missing path",
			req: CreateGitWorkflow{
				CommitHash: "8458fd753f9fde51882414564c20df6d4c34a90e",
			},
			wantErr: errors.New("path is required"),
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

func TestCreateTargetValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateTarget
		types   []string
		wantErr error
	}{
		{
			name: "valid minimal",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "aws_account",
			},
		},
		{
			name: "valid full",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					RoleArn:        "arn:aws:iam::012345678901:role/test-role",
					PolicyDocument: "{ \"Version\": \"2012-10-17\", \"Statement\": [ { \"Effect\": \"Allow\", \"Action\": \"s3:ListBuckets\", \"Resource\": \"*\" } ] }",
					PolicyArns: []string{
						"arn:aws:iam::012345678901:policy/test-policy-1",
						"arn:aws:iam::012345678901:policy/test-policy-2",
						"arn:aws:iam::012345678901:policy/test-policy-3",
						"arn:aws:iam::012345678901:policy/test-policy-4",
						"arn:aws:iam::012345678901:policy/test-policy-5",
					},
				},
				Type: "aws_account",
			},
		},
		{
			name: "missing name",
			req: CreateTarget{
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("name is required"),
		},
		{
			name: "name must be alphanumeric underscore",
			req: CreateTarget{
				Name: "this-is-invalid",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("name must be alphanumeric underscore"),
		},
		{
			name: "too short name",
			req: CreateTarget{
				Name: "abc",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("name must be between 4 and 32 characters"),
		},
		{
			name: "too long name",
			req: CreateTarget{
				Name: "a12345678901234567890123456789012",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("name must be between 4 and 32 characters"),
		},
		{
			name: "missing type",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				},
			},
			wantErr: errors.New("type is required"),
		},
		{
			name: "invalid type",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "bad",
			},
			wantErr: errors.New("type must be one of 'aws_account'"),
		},
		{
			name: "missing credential_type",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					RoleArn: "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("credential_type is required"),
		},
		{
			name: "invalid credential_type",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					CredentialType: "bad",
					RoleArn:        "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("credential_type must be one of 'assumed_role'"),
		},
		{
			name: "missing role_arn",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("role_arn is required"),
		},
		{
			name: "role_arn must be an arn",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					RoleArn:        "not-an-arn",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("role_arn must be a valid arn"),
		},
		{
			name: "too many policy arns",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					PolicyArns: []string{
						"arn:aws:iam::012345678901:policy/test-policy-1",
						"arn:aws:iam::012345678901:policy/test-policy-2",
						"arn:aws:iam::012345678901:policy/test-policy-3",
						"arn:aws:iam::012345678901:policy/test-policy-4",
						"arn:aws:iam::012345678901:policy/test-policy-5",
						"arn:aws:iam::012345678901:policy/test-policy-6",
					},
					RoleArn: "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("policy_arns cannot be more than 5"),
		},
		{
			name: "policy arns must be valid",
			req: CreateTarget{
				Name: "target1",
				Properties: TargetProperties{
					CredentialType: "assumed_role",
					PolicyArns: []string{
						"arn:aws:iam::012345678901:policy/test-policy-1",
						"not-an-arn",
					},
					RoleArn: "arn:aws:iam::012345678901:role/test-role",
				},
				Type: "aws_account",
			},
			wantErr: errors.New("policy_arns contains an invalid arn"),
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

func TestTargetOperationValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     TargetOperation
		types   []string
		wantErr error
	}{
		{
			name: "valid",
			req: TargetOperation{
				Path: "./manifest.yaml",
				SHA:  "8458fd753f9fde51882414564c20df6d4c34a90e",
				Type: "diff",
			},
		},
		{
			name: "missing commit hash",
			req: TargetOperation{
				Path: "./manifest.yaml",
				Type: "diff",
			},
			wantErr: errors.New("sha is required"),
		},
		{
			name: "commit hash must be alphanumeric",
			req: TargetOperation{
				SHA:  "8--",
				Path: "./manifest.yaml",
				Type: "diff",
			},
			wantErr: errors.New("sha must be alphanumeric"),
		},
		{
			name: "missing path",
			req: TargetOperation{
				SHA:  "8458fd753f9fde51882414564c20df6d4c34a90e",
				Type: "diff",
			},
			wantErr: errors.New("path is required"),
		},
		{
			name: "missing type",
			req: TargetOperation{
				SHA:  "8458fd753f9fde51882414564c20df6d4c34a90e",
				Path: "./manifest.yaml",
			},
			wantErr: errors.New("type is required"),
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

func TestCreateProjectValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateProject
		types   []string
		wantErr error
	}{
		{
			name: "valid",
			req: CreateProject{
				Name:       "project1",
				Repository: "https://github.com/argoproj-labs/argo-cloudops.git",
			},
		},
		{
			name: "missing name",
			req: CreateProject{
				Repository: "https://github.com/argoproj-labs/argo-cloudops.git",
			},
			wantErr: errors.New("name is required"),
		},
		{
			name: "name must be alphanumeric",
			req: CreateProject{
				Name:       "this-is-invalid",
				Repository: "https://github.com/argoproj-labs/argo-cloudops.git",
			},
			wantErr: errors.New("name must be alphanumeric"),
		},
		{
			name: "too short name",
			req: CreateProject{
				Name:       "abc",
				Repository: "https://github.com/argoproj-labs/argo-cloudops.git",
			},
			wantErr: errors.New("name must be between 4 and 32 characters"),
		},
		{
			name: "too long name",
			req: CreateProject{
				Name:       "a12345678901234567890123456789012",
				Repository: "https://github.com/argoproj-labs/argo-cloudops.git",
			},
			wantErr: errors.New("name must be between 4 and 32 characters"),
		},
		{
			name: "missing repository",
			req: CreateProject{
				Name: "project1",
			},
			wantErr: errors.New("repository is required"),
		},
		{
			name: "invalid repository",
			req: CreateProject{
				Name:       "project1",
				Repository: "invalid-repo",
			},
			wantErr: errors.New("repository must be a git uri"),
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
