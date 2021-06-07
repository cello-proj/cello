package credentials

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	vault "github.com/hashicorp/vault/api"
)

func TestVaultProjectExists(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		exists    bool
		vaultErr  error
		errResult error
	}{
		{
			name:      "get project success",
			path:      "test-path",
			exists:    true,
			errResult: nil,
		},
		{
			name:      "get project not found",
			path:      "test-path",
			exists:    false,
			vaultErr:  ErrNotFound,
			errResult: fmt.Errorf("vault get project error: %v", ErrNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			v := VaultProvider{
				VaultLogicalSvc: &mockVaultLogical{err: tt.vaultErr},
			}

			status, err := v.ProjectExists(tt.path)
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

			if !cmp.Equal(status, tt.exists) {
				t.Errorf("\nwant: %v\n got: %v", tt.exists, status)
			}
		})
	}
}

type mockVaultLogical struct {
	vault.Logical
	err error
}

func (m mockVaultLogical) Read(path string) (*vault.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &vault.Secret{}, nil
}
