package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionsToMap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "input is not empty string",
			input: "sslrootcert=rds-ca.pem sslmode=verify-full",
			expected: map[string]string{
				"sslrootcert": "rds-ca.pem",
				"sslmode":     "verify-full",
			},
		},
		{
			name:     "input is empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "option has multiple equal signs",
			input: "sslrootcert=rds-ca.pem sslmode=verify-full option=value=value",
			expected: map[string]string{
				"sslrootcert": "rds-ca.pem",
				"sslmode":     "verify-full",
				"option":      "value=value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OptionsToMap(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
