package credentials

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	vault "github.com/hashicorp/vault/api"
)

var testErr = fmt.Errorf("error")

func TestVaultCreateProject(t *testing.T) {
	tests := []struct {
		name           string
		admin          bool
		expectedRole   string
		expectedSecret string
		vaultErr       error
		errResult      bool
	}{
		{
			name:           "create project success",
			admin:          true,
			expectedSecret: "test-secret",
			expectedRole:   "test-role",
		},
		{
			name:      "create project admin error",
			admin:     false,
			errResult: true,
		},
		{
			name:      "create prolicy error",
			admin:     true,
			vaultErr:  testErr,
			errResult: true,
		},
		{
			name:      "create project error",
			admin:     true,
			vaultErr:  testErr,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin == true {
				role = "admin"
			}
			v := VaultProvider{
				roleID: role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr, data: map[string]interface{}{
					"secret_id": tt.expectedSecret,
					"role_id":   tt.expectedRole,
				}},
				vaultSysSvc: &mockVaultSys{},
			}

			roleID, secretID, err := v.CreateProject("testProject")
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
			vaultErr:  testErr,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin == true {
				role = "admin"
			}
			v := VaultProvider{
				roleID:          role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
			}

			err := v.CreateTarget("test", CreateTargetRequest{})
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
			vaultErr:  testErr,
			errResult: true,
		},
		{
			name:           "delete project policy error",
			admin:          true,
			vaultPolicyErr: testErr,
			errResult:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin == true {
				role = "admin"
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
			vaultErr:  testErr,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin == true {
				role = "admin"
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
		name               string
		admin              bool
		expectedProperties TargetProperties
		vaultErr           error
		errResult          bool
	}{
		{
			name:  "get target success",
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
			vaultErr:  testErr,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin == true {
				role = "admin"
			}
			v := VaultProvider{
				roleID: role,
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr, data: map[string]interface{}{
					"role_arns":       []interface{}{"test-role-arn"},
					"policy_arns":     []interface{}{"test-policy-arn"},
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
			vaultErr:  testErr,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin == true {
				role = "admin"
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
			vaultErr:  testErr,
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var role = "testRole"
			if tt.admin == true {
				role = "admin"
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
		errResult bool
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
			errResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			v := VaultProvider{
				vaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
			}

			status, err := v.ProjectExists(tt.path)
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}

				if !cmp.Equal(status, tt.exists) {
					t.Errorf("\nwant: %v\n got: %v", tt.exists, status)
				}
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
