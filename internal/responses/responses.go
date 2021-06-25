package responses

type GetProject struct {
	Name string `json:"name"`
}

type GetTarget struct {
	Name string `json:"name"`
}

// Target properties for target responses.
type TargetProperties struct {
	CredentialType string   `json:"credential_type"`
	PolicyArns     []string `json:"policy_arns"`
	RoleArn        string   `json:"role_arn"`
}
