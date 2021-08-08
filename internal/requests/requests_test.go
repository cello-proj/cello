package requests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestCreateWorkflowValidateParameters(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]string
		wantErr    error
	}{
		{
			name: "valid",
			parameters: map[string]string{
				"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				"pre_container_image_uri":     "argoproj-labs/argo-cloudops-pre",
			},
		},
		{
			name: "valid_no_pre_image",
			parameters: map[string]string{
				"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
			},
		},
		{
			name: "invalid_missing_exec",
			parameters: map[string]string{
				"pre_container_image_uri": "argoproj-labs/argo-cloudops-pre",
			},
			wantErr: errors.New("parameter execute_container_image_uri is required"),
		},
		{
			name: "invalid_bad_exec",
			parameters: map[string]string{
				"execute_container_image_uri": "./foo/bar",
			},
			wantErr: errors.New("parameter execute_container_image_uri must be a valid container uri"),
		},
		{
			name: "invalid_bad_pre",
			parameters: map[string]string{
				"execute_container_image_uri": "argoproj-labs/argo-cloudops-exec",
				"pre_container_image_uri":     "./foo/bar",
			},
			wantErr: errors.New("parameter pre_container_image_uri must be a valid container uri"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateWorkflow{
				Parameters: tt.parameters,
			}
			assert.Equal(t, tt.wantErr, req.validateParameters())
		})
	}
}

func TestCreateWorkflowValidateArguments(t *testing.T) {
	tests := []struct {
		name      string
		arguments map[string][]string
		wantErr   error
	}{
		{
			name: "valid",
			arguments: map[string][]string{
				"execute": {"--foo", "--bar"},
				"init":    {"--baz", "blah"},
			},
		},
		{
			name:      "valid_empty",
			arguments: map[string][]string{},
		},
		{
			name: "invalid_too_many",
			arguments: map[string][]string{
				"execute": {"--foo", "--bar"},
				"init":    {"--baz", "blah"},
				"other":   {"--not", "valid"},
			},
			wantErr: errors.New("arguments must be one of 'execute init'"),
		},
		{
			name: "invalid_not_execute_or_init",
			arguments: map[string][]string{
				"execute": {"--foo", "--bar"},
				"other":   {"--not", "valid"},
			},
			wantErr: errors.New("arguments must be one of 'execute init'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateWorkflow{
				Arguments: tt.arguments,
			}
			assert.Equal(t, tt.wantErr, req.validateArguments())
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
