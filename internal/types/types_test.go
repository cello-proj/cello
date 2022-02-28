package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTargetPropertiesValidate(t *testing.T) {
	tests := []struct {
		name       string
		properties TargetProperties
		types      []string
		wantErr    error
	}{
		{
			name: "valid minimal",
			properties: TargetProperties{
				CredentialType: "assumed_role",
				RoleArn:        "arn:aws:iam::012345678901:role/test-role",
			},
		},
		{
			name: "valid full",
			properties: TargetProperties{
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
		},
		{
			name: "role_arn must be an arn",
			properties: TargetProperties{
				CredentialType: "assumed_role",
				RoleArn:        "not-an-arn",
			},
			wantErr: errors.New("role_arn must be a valid arn"),
		},
		{
			name: "too many policy arns",
			properties: TargetProperties{
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
			wantErr: errors.New("policy_arns cannot be more than 5"),
		},
		{
			name: "policy arns must be valid",
			properties: TargetProperties{
				CredentialType: "assumed_role",
				PolicyArns: []string{
					"arn:aws:iam::012345678901:policy/test-policy-1",
					"not-an-arn",
				},
				RoleArn: "arn:aws:iam::012345678901:role/test-role",
			},
			wantErr: errors.New("policy_arns contains an invalid arn"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr != nil {
				assert.EqualError(t, tt.properties.Validate(), tt.wantErr.Error())
			} else {
				assert.Equal(t, tt.wantErr, tt.properties.Validate())
			}
		})
	}
}

func TestCTargetValidate(t *testing.T) {
	tests := []struct {
		name    string
		target  Target
		types   []string
		wantErr error
	}{
		{
			name: "valid minimal",
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
			target: Target{
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
				assert.EqualError(t, tt.target.Validate(), tt.wantErr.Error())
			} else {
				assert.Equal(t, tt.wantErr, tt.target.Validate())
			}
		})
	}
}
