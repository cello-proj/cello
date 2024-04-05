package db

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSQLClientWithOptions(t *testing.T) {
	tests := []struct {
		name      string
		dbProps   map[string]string
		dbOptions map[string]string
	}{
		{
			name: "sqlClientWithOptions",
			dbProps: map[string]string{
				"host":     "localhost",
				"database": "argocloudops",
				"user":     "argocd",
				"password": "argocd_password",
				"options":  "sslrootcert=rds-ca.pem sslmode=verify-full",
			},
			dbOptions: map[string]string{
				"sslrootcert": "rds-ca.pem",
				"sslmode":     "verify-full",
			},
		},
		{
			name: "sqlClientWithoutOptions",
			dbProps: map[string]string{
				"host":     "localhost",
				"database": "argocloudops",
				"user":     "argocd",
				"password": "argocd_password",
				"options":  "",
			},
			dbOptions: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewSQLClient(
				tt.dbProps["host"],
				tt.dbProps["database"],
				tt.dbProps["user"],
				tt.dbProps["password"],
				tt.dbProps["options"],
			)

			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(db.options, tt.dbOptions))
		})
	}
}
