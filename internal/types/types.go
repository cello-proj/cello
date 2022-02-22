package types

type Target struct {
	Name       string           `json:"name" valid:"required~name is required,alphanumunderscore~name must be alphanumeric underscore,stringlength(4|32)~name must be between 4 and 32 characters"`
	Properties TargetProperties `json:"properties"`
	Type       string           `json:"type" valid:"required~type is required"`
}

// TargetProperties for target
type TargetProperties struct {
	CredentialType string   `json:"credential_type"`
	PolicyArns     []string `json:"policy_arns"`
	PolicyDocument string   `json:"policy_document"`
	RoleArn        string   `json:"role_arn"`
}
