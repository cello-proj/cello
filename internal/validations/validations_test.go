package validations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateIsAlphaNumericUnderscore(t *testing.T) {
	type testStruct struct {
		Test string `validate:"is_alphanumunderscore"`
	}

	tests := []struct {
		name       string
		testString string
		wantErr    error
	}{
		{
			name:       "valid alpha num underscore",
			testString: "abcd1234____",
		},
		{
			name:       "invalid alpha num underscore characters",
			testString: "--[[]]  ",
			wantErr:    fmt.Errorf("value '--[[]]  ' is invalid, must only contain alpha numberic underscore characters"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidationStruct := testStruct{Test: tt.testString}
			err := ValidateStruct(&testValidationStruct)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateContainerImages(t *testing.T) {
	type testStruct struct {
		Parameters map[string]string `validate:"is_valid_execute_container_image"`
	}

	tests := []struct {
		name                  string
		testString            string
		noExecuteContainerKey bool
		wantErr               error
	}{
		{
			name:       "valid execute container image",
			testString: "argocloudops/argo-cloudops-cdk:1.87.1",
		},
		{
			name:       "invalid execute container image",
			testString: "()argocloudops  -- /argo-cloudops-cdk:1.87.1",
			wantErr:    fmt.Errorf("'Parameters' value 'map[execute_container_image_uri:()argocloudops  -- /argo-cloudops-cdk:1.87.1]' is an invalid container uri"),
		},
		{
			name:    "no image provided",
			wantErr: fmt.Errorf("'Parameters' value 'map[execute_container_image_uri:]' is an invalid container uri"),
		},
		{
			name:                  "no execute container key",
			noExecuteContainerKey: true,
			wantErr:               fmt.Errorf("'Parameters' value 'map[]' is an invalid container uri"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testValidationStruct testStruct
			if !tt.noExecuteContainerKey {
				testValidationStruct.Parameters = make(map[string]string)
				testValidationStruct.Parameters["execute_container_image_uri"] = tt.testString
			}
			err := ValidateStruct(&testValidationStruct)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestPreContainerImages(t *testing.T) {
	type testStruct struct {
		Parameters map[string]string `validate:"is_valid_precontainer_image"`
	}

	tests := []struct {
		name              string
		testString        string
		noPreContainerKey bool
		wantErr           error
	}{
		{
			name:       "valid pre container image",
			testString: "argocloudops/argo-cloudops-cdk:1.87.1",
		},
		{
			name:       "invalid pre container image",
			testString: "()argocloudops  -- /argo-cloudops-cdk:1.87.1",
			wantErr:    fmt.Errorf("'Parameters' value 'map[pre_container_image_uri:()argocloudops  -- /argo-cloudops-cdk:1.87.1]' is an invalid container uri"),
		},
		{
			name:    "no image provided",
			wantErr: fmt.Errorf("'Parameters' value 'map[pre_container_image_uri:]' is an invalid container uri"),
		},
		{
			name:              "no provided precontainer key, optional no error",
			noPreContainerKey: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testValidationStruct testStruct
			if !tt.noPreContainerKey {
				testValidationStruct.Parameters = make(map[string]string)
				testValidationStruct.Parameters["pre_container_image_uri"] = tt.testString
			}
			err := ValidateStruct(&testValidationStruct)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidTargetType(t *testing.T) {
	type testStruct struct {
		Type string `validate:"is_valid_target_type"`
	}

	tests := []struct {
		name       string
		testString string
		wantErr    error
	}{
		{
			name:       "valid target type",
			testString: "aws_account",
		},
		{
			name:       "invalid target type",
			testString: "not_aws_account",
			wantErr:    fmt.Errorf("'Type' value 'not_aws_account' is invalid, types supported:'aws_account'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidationStruct := testStruct{Type: tt.testString}
			err := ValidateStruct(testValidationStruct)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidArgument(t *testing.T) {
	type testStruct struct {
		Arguments map[string][]string `validate:"is_valid_argument"`
	}

	tests := []struct {
		name              string
		testString        string
		noPreContainerKey bool
		wantErr           error
	}{
		{
			name:       "valid argument init",
			testString: "init",
		},
		{
			name:       "valid argument execute",
			testString: "execute",
		},
		{
			name:       "invalid argument",
			testString: "exec",
			wantErr:    fmt.Errorf("'Arguments' value 'map[exec:[foo]]' is an invalid argument"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testValidationStruct testStruct
			if !tt.noPreContainerKey {
				testValidationStruct.Arguments = make(map[string][]string)
				testValidationStruct.Arguments[tt.testString] = []string{"foo"}
			}
			err := ValidateStruct(&testValidationStruct)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidArn(t *testing.T) {
	type testStruct struct {
		Arn string `validate:"is_arn"`
	}

	tests := []struct {
		name       string
		testString string
		wantErr    error
	}{
		{
			name:       "valid arn",
			testString: "arn:aws:iam::012345678901:policy/test-policy",
		},
		{
			name:       "invalid arn",
			testString: "invalid-arn",
			wantErr:    fmt.Errorf("'Arn' value 'invalid-arn' is not a valid arn"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidationStruct := testStruct{Arn: tt.testString}
			err := ValidateStruct(testValidationStruct)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateVar(t *testing.T) {
	tests := []struct {
		name        string
		testString  string
		validString string
		wantErr     error
	}{
		{
			name:        "valid var",
			testString:  "good",
			validString: "good",
		},
		{
			name:        "invalid var",
			testString:  "bad",
			validString: "good",
			wantErr:     fmt.Errorf("'validate var' 'good'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar("validate var", tt.testString, fmt.Sprintf("eq=%s", tt.validString))
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateStructError(t *testing.T) {
	tests := []struct {
		name             string
		validationStruct interface{}
		errResult        bool
		wantErr          error
	}{
		{
			name: "no error",
			validationStruct: struct {
				Type string `validate:"is_valid_target_type"`
			}{"aws_account"},
		},
		{
			name: "invalid target type string",
			validationStruct: struct {
				Type string `validate:"is_valid_target_type"`
			}{"bad"},
			errResult: true,
			wantErr:   fmt.Errorf("'Type' value 'bad' is invalid, types supported:'aws_account'"),
		},
		{
			name: "non arn provider to is arn validation",
			validationStruct: struct {
				Arn string `validate:"is_arn"`
			}{"bad"},
			errResult: true,
			wantErr:   fmt.Errorf("'Arn' value 'bad' is not a valid arn"),
		},
		{
			name: "bad values (-)'s in alpha numeric underscore validation",
			validationStruct: struct {
				TargetName string `validate:"is_alphanumunderscore"`
			}{"bad-value"},
			errResult: true,
			wantErr:   fmt.Errorf("value 'bad-value' is invalid, must only contain alpha numberic underscore characters"),
		},
		{
			name: "bad execute container uri",
			validationStruct: struct {
				Parameters map[string]string `validate:"is_valid_execute_container_image"`
			}{map[string]string{
				"execute_container_image_uri": "bad()",
			}},
			errResult: true,
			wantErr:   fmt.Errorf("'Parameters' value 'map[execute_container_image_uri:bad()]' is an invalid container uri"),
		},
		{
			name: "bad pre container uri",
			validationStruct: struct {
				Parameters map[string]string `validate:"is_valid_precontainer_image"`
			}{map[string]string{
				"pre_container_image_uri": "bad()",
			}},
			errResult: true,
			wantErr:   fmt.Errorf("'Parameters' value 'map[pre_container_image_uri:bad()]' is an invalid container uri"),
		},
		{
			name: "invalid argument",
			validationStruct: struct {
				Argument map[string][]string `validate:"is_valid_argument"`
			}{map[string][]string{
				"exec": {""}}},
			errResult: true,
			wantErr:   fmt.Errorf("'Argument' value 'map[exec:[]]' is an invalid argument"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidationStruct := tt.validationStruct
			err := ValidateStruct(testValidationStruct)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateVarErrors(t *testing.T) {
	tests := []struct {
		name        string
		testString  string
		validString string
		wantErr     error
	}{
		{
			name:        "valid var",
			testString:  "good",
			validString: "good",
		},
		{
			name:        "invalid var expected error",
			testString:  "bad",
			validString: "good",
			wantErr:     fmt.Errorf("'validate var' 'good'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar("validate var", tt.testString, fmt.Sprintf("eq=%s", tt.validString))
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
