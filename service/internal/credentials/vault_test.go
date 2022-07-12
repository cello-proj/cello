package credentials

import (
	"fmt"
	"testing"

	"github.com/cello-proj/cello/internal/types"

	"github.com/google/go-cmp/cmp"
	vault "github.com/hashicorp/vault/api"
)

var errTest = fmt.Errorf("error")

func TestVaultCreateProject(t *testing.T) {
	tests := []struct {
		name                   string
		admin                  bool
		expectedRole           string
		expectedSecret         string
		expectedSecretAccessor string
		vaultErr               error
		errResult              bool
	}{
		{
			name:                   "create project success",
			admin:                  true,
			expectedSecret:         "test-secret",
			expectedSecretAccessor: "test-secret-accessor",
			expectedRole:           "test-role",
		},
		{
			name:      "create project admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "create project error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin {
				role = authorizationKeyAdmin
			}
			v := VaultProvider{
				roleID: role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr, data: map[string]interface{}{
					"secret_id":          tt.expectedSecret,
					"secret_id_accessor": tt.expectedSecretAccessor,
					"role_id":            tt.expectedRole,
				}},
				vaultSysSvc: &mockVaultSys{},
			}

			roleID, secretID, _, err := v.CreateProject("testProject")
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
				if !cmp.Equal(roleID, tt.expectedRole) {
					t.Errorf("\nwant: %v\n got: %v", tt.expectedRole, roleID)
				}
				if !cmp.Equal(secretID, tt.expectedSecret) {
					t.Errorf("\nwant: %v\n got: %v", tt.expectedSecret, secretID)
				}
			}

		})
	}
}

func TestVaultCreateTarget(t *testing.T) {
	tests := []struct {
		name      string
		admin     bool
		vaultErr  error
		errResult bool
	}{
		{
			name:  "create target success",
			admin: true,
		},
		{
			name:      "create target admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "create target error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin {
				role = authorizationKeyAdmin
			}
			v := VaultProvider{
				roleID:          role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
			}

			err := v.CreateTarget("test", types.Target{})
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
			}
		})
	}
}

func TestVaultUpdateTarget(t *testing.T) {
	tests := []struct {
		name      string
		admin     bool
		vaultErr  error
		errResult bool
	}{
		{
			name:  "update target success",
			admin: true,
		},
		{
			name:      "update target admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "update target error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin {
				role = authorizationKeyAdmin
			}
			v := VaultProvider{
				roleID:          role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
			}

			err := v.UpdateTarget("test", types.Target{})
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
			}
		})
	}
}

func TestVaultDeleteProject(t *testing.T) {
	tests := []struct {
		name           string
		admin          bool
		vaultErr       error
		vaultPolicyErr error
		errResult      bool
	}{
		{
			name:  "delete project success",
			admin: true,
		},
		{
			name:      "delete project admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "delete project error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
		{
			name:           "delete project policy error",
			admin:          true,
			vaultPolicyErr: errTest,
			errResult:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin {
				role = authorizationKeyAdmin
			}
			v := VaultProvider{
				roleID:          role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
				vaultSysSvc:     &mockVaultSys{err: tt.vaultPolicyErr},
			}

			err := v.DeleteProject("testProject")
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
			}
		})
	}
}

func TestVaultDeleteTarget(t *testing.T) {
	tests := []struct {
		name      string
		admin     bool
		vaultErr  error
		errResult bool
	}{
		{
			name:  "delete target success",
			admin: true,
		},
		{
			name:      "delete target admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "delete target error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin {
				role = authorizationKeyAdmin
			}
			v := VaultProvider{
				roleID:          role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
			}

			err := v.DeleteTarget("testProject", "testTarget")
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
			}
		})
	}
}

func TestVaultGetTarget(t *testing.T) {
	tests := []struct {
		name      string
		admin     bool
		vaultErr  error
		errResult bool
	}{
		{
			name:  "get target success",
			admin: true,
		},
		{
			name:      "get target admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "get target error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin {
				role = authorizationKeyAdmin
			}
			v := VaultProvider{
				roleID: role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr, data: map[string]interface{}{
					"role_arns":       []interface{}{"test-role-arn"},
					"policy_arns":     []interface{}{"test-policy-arn"},
					"policy_document": `{ "Version": "2012-10-17", "Statement": [ { "Effect": "Allow", "Action": "s3:ListBuckets", "Resource": "*" } ] }`,
					"credential_type": "test-cred-type",
				}},
			}

			_, err := v.GetTarget("testProject", "testTarget")
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
			}
		})
	}
}

func TestVaultGetToken(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		admin     bool
		vaultErr  error
		errResult bool
	}{
		{
			name:  "get token success",
			token: "secretToken",
		},
		{
			name:      "get token admin error",
			admin:     true,
			errResult: true,
		},
		{
			name:      "get token error",
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin {
				role = authorizationKeyAdmin
			}
			v := VaultProvider{
				roleID:          role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr, token: tt.token},
			}

			token, err := v.GetToken()
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
				if !cmp.Equal(token, tt.token) {
					t.Errorf("\nwant: %v\n got: %v", tt.token, token)
				}
			}
		})
	}
}

func TestVaultListTargets(t *testing.T) {
	tests := []struct {
		name            string
		admin           bool
		targets         []string
		expectedTargets []string
		vaultErr        error
		errResult       bool
	}{
		{
			name:    "list target success",
			admin:   true,
			targets: []string{"target1", "target2"},
		},
		{
			name:      "list target admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "list target error",
			admin:     true,
			vaultErr:  errTest,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin {
				role = authorizationKeyAdmin
			}
			var testTargets []interface{}
			for _, i := range tt.targets {
				testTargets = append(testTargets, fmt.Sprintf("argo-cloudops-projects-test-target-%s", i))
			}
			v := VaultProvider{
				roleID: role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr, data: map[string]interface{}{
					"keys": testTargets,
				}},
			}

			targets, err := v.ListTargets("test")
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
				if !cmp.Equal(targets, tt.targets) {
					t.Errorf("\nwant: %v\n got: %v", tt.targets, targets)
				}
			}
		})
	}
}

func TestVaultProjectExists(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		exists    bool
		vaultErr  error
		expectErr bool
	}{
		{
			name:   "get project success",
			path:   "test-path",
			exists: true,
		},
		{
			name:      "get project not found",
			path:      "test-path",
			exists:    false,
			vaultErr:  ErrNotFound,
			expectErr: false,
		},
		{
			name:      "vault error",
			path:      "test-path",
			exists:    false,
			vaultErr:  errTest,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			v := VaultProvider{
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
			}

			status, err := v.ProjectExists(tt.path)
			if err != nil {
				if !tt.expectErr {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.expectErr {
					t.Errorf("\nexpected error")
				}

				if !cmp.Equal(status, tt.exists) {
					t.Errorf("\nwant: %v\n got: %v", tt.exists, status)
				}
			}
		})
	}
}

func TestValidateAuthorizedAdmin(t *testing.T) {
	tests := []struct {
		name        string
		admin       bool
		validSecret bool
		expectErr   bool
	}{
		{
			name:        "is authorized admin",
			admin:       true,
			validSecret: true,
			expectErr:   false,
		},
		{
			name:        "isn't admin, with valid secret",
			admin:       false,
			validSecret: true,
			expectErr:   true,
		},
		{
			name:        "is admin, with invalid secret",
			admin:       true,
			validSecret: false,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var key = "test"
			if tt.admin {
				key = authorizationKeyAdmin
			}
			var secret = "invalidSecret"
			if tt.validSecret {
				secret = "validSecret"
			}
			a := Authorization{Provider: "vault", Key: key, Secret: secret}
			err := a.Validate(a.ValidateAuthorizedAdmin("validSecret"))
			if err != nil != tt.expectErr {
				t.Errorf("\nwant error: %v\n got error: %v", tt.expectErr, err != nil)
			}
		})
	}
}

func TestNewAuthorization(t *testing.T) {
	tests := []struct {
		name         string
		header       string
		expectedAuth *Authorization
		expectErr    bool
	}{
		{
			name:         "valid authorization header",
			header:       "vault:testkey:testsecret",
			expectedAuth: &Authorization{"vault", "testkey", "testsecret"},
			expectErr:    false,
		},
		{
			name:      "invalid authorization header",
			header:    "vault:testbad",
			expectErr: true,
		},
		{
			name:      "authorization header empty",
			header:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewAuthorization(tt.header)
			if err != nil != tt.expectErr {
				t.Errorf("\nwant error: %v\n got error: %v", tt.expectErr, err != nil)
			}
			if !cmp.Equal(a, tt.expectedAuth) {
				t.Errorf("\nwant auth: %v\n got: %v", tt.expectedAuth, a)
			}
		})
	}
}

type mockVaultLogical struct {
	vault.Logical
	data  map[string]interface{}
	token string
	err   error
}

func (m mockVaultLogical) Read(path string) (*vault.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &vault.Secret{Data: m.data}, nil
}

func (m mockVaultLogical) List(path string) (*vault.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &vault.Secret{Data: m.data}, nil
}

func (m mockVaultLogical) Write(path string, data map[string]interface{}) (*vault.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &vault.Secret{Data: m.data, Auth: &vault.SecretAuth{ClientToken: m.token}}, nil
}

func (m mockVaultLogical) Delete(path string) (*vault.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &vault.Secret{}, nil
}

type mockVaultSys struct {
	vault.Sys
	err error
}

func (m mockVaultSys) PutPolicy(name, rules string) error {
	return m.err
}

func (m mockVaultSys) DeletePolicy(name string) error {
	return m.err
}
