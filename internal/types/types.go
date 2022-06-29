package types

import (
	"errors"

	"github.com/cello-proj/cello/internal/validations"
)

type Target struct {
	Name       string           `json:"name" valid:"required~name is required,alphanumunderscore~name must be alphanumeric underscore,stringlength(4|32)~name must be between 4 and 32 characters"`
	Properties TargetProperties `json:"properties"`
	Type       string           `json:"type" valid:"required~type is required"`
}

// TargetProperties for target
type TargetProperties struct {
	CredentialType string   `json:"credential_type" valid:"required~credential_type is required"`
	PolicyArns     []string `json:"policy_arns"`
	PolicyDocument string   `json:"policy_document"`
	RoleArn        string   `json:"role_arn" valid:"required~role_arn is required"`
}

// Validate validates Target.
func (target Target) Validate() error {
	v := []func() error{
		func() error { return validations.ValidateStruct(target) },
		func() error {
			if target.Type != "aws_account" {
				return errors.New("type must be one of 'aws_account'")
			}
			return nil
		},
		target.Properties.Validate,
	}

	return validations.Validate(v...)
}

// Validate validates TargetProperties.
func (properties TargetProperties) Validate() error {
	v := []func() error{
		func() error { return validations.ValidateStruct(properties) },
		func() error {
			if properties.CredentialType != "assumed_role" {
				return errors.New("credential_type must be one of 'assumed_role'")
			}

			if !validations.IsValidARN(properties.RoleArn) {
				return errors.New("role_arn must be a valid arn")
			}

			if len(properties.PolicyArns) > 5 {
				return errors.New("policy_arns cannot be more than 5")
			}

			for _, arn := range properties.PolicyArns {
				if !validations.IsValidARN(arn) {
					return errors.New("policy_arns contains an invalid arn")
				}
			}
			return nil
		},
	}

	return validations.Validate(v...)
}

type ProjectToken struct {
	ID string `json:"token_id"`
}
