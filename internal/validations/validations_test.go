package validations

import (
	"fmt"
	"testing"
)

func TestValidateIsAlphaNumericUnderscore(t *testing.T) {
	type testStruct struct {
		Test string `validate:"alphanumunderscore"`
	}

	tests := []struct {
		name       string
		testString string
		errResult  bool
	}{
		{
			name:       "valid alpha num underscore",
			testString: "abcd1234____",
		},
		{
			name:       "invalid alpha num underscore characters",
			testString: "--[[]]  ",
			errResult:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidationStruct := testStruct{Test: tt.testString}
			err := ValidateStruct(&testValidationStruct)
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateContainerImages(t *testing.T) {
	type testStruct struct {
		Test map[string]string `validate:"valid_execute_container_image"`
	}

	tests := []struct {
		name                  string
		testString            string
		noExecuteContainerKey bool
		errResult             bool
	}{
		{
			name:       "valid execute container image",
			testString: "argocloudops/argo-cloudops-cdk:1.87.1",
		},
		{
			name:       "invalid execute container image",
			testString: "()argocloudops  -- /argo-cloudops-cdk:1.87.1",
			errResult:  true,
		},
		{
			name:      "no image provided",
			errResult: true,
		},
		{
			name:                  "no execute container key",
			noExecuteContainerKey: true,
			errResult:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testValidationStruct testStruct
			if !tt.noExecuteContainerKey {
				testValidationStruct.Test = make(map[string]string)
				testValidationStruct.Test["execute_container_image_uri"] = tt.testString
			}
			err := ValidateStruct(&testValidationStruct)
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error, got: %v", err)
				}
			}
		})
	}
}

func TestPreContainerImages(t *testing.T) {
	type testStruct struct {
		Test map[string]string `validate:"valid_precontainer_image"`
	}

	tests := []struct {
		name              string
		testString        string
		noPreContainerKey bool
		errResult         bool
	}{
		{
			name:       "valid pre container image",
			testString: "argocloudops/argo-cloudops-cdk:1.87.1",
		},
		{
			name:       "invalid pre container image",
			testString: "()argocloudops  -- /argo-cloudops-cdk:1.87.1",
			errResult:  true,
		},
		{
			name:      "no image provided",
			errResult: true,
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
				testValidationStruct.Test = make(map[string]string)
				testValidationStruct.Test["pre_container_image_uri"] = tt.testString
			}
			err := ValidateStruct(&testValidationStruct)
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error, got: %v", err)
				}
			}
		})
	}
}

func TestValidTargetType(t *testing.T) {
	type testStruct struct {
		Test string `validate:"valid_target_type"`
	}

	tests := []struct {
		name       string
		testString string
		errResult  bool
	}{
		{
			name:       "valid target type",
			testString: "aws_account",
		},
		{
			name:       "invalid target type",
			testString: "not_aws_account",
			errResult:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidationStruct := testStruct{Test: tt.testString}
			err := ValidateStruct(testValidationStruct)
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error, got: %v", err)
				}
			}
		})
	}
}

func TestValidArgument(t *testing.T) {
	type testStruct struct {
		Test map[string][]string `validate:"valid_argument"`
	}

	tests := []struct {
		name              string
		testString        string
		noPreContainerKey bool
		errResult         bool
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
			errResult:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testValidationStruct testStruct
			if !tt.noPreContainerKey {
				testValidationStruct.Test = make(map[string][]string)
				testValidationStruct.Test[tt.testString] = []string{"foo"}
			}
			err := ValidateStruct(&testValidationStruct)
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error, got: %v", err)
				}
			}
		})
	}
}

func TestValidArn(t *testing.T) {
	type testStruct struct {
		Test string `validate:"is_arn"`
	}

	tests := []struct {
		name       string
		testString string
		errResult  bool
	}{
		{
			name:       "valid arn",
			testString: "arn:aws:iam::012345678901:policy/test-policy",
		},
		{
			name:       "invalid arn",
			testString: "invalid-arn",
			errResult:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidationStruct := testStruct{Test: tt.testString}
			err := ValidateStruct(testValidationStruct)
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateVar(t *testing.T) {
	tests := []struct {
		name        string
		testString  string
		validString string
		errResult   bool
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
			errResult:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar("validate var", tt.testString, fmt.Sprintf("eq=%s", tt.validString))
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateStructError(t *testing.T) {
	tests := []struct {
		name             string
		validationStruct interface{}
		errResult        bool
		errString        string
	}{
		{
			name: "no error",
			validationStruct: struct {
				Test string `validate:"valid_target_type"`
			}{"aws_account"},
		},
		{
			name: "invalid target type string",
			validationStruct: struct {
				Test string `validate:"valid_target_type"`
			}{"bad"},
			errResult: true,
			errString: "failed validation check for 'valid_target_type', value 'bad' is invalid, types supported:'aws_account'",
		},
		{
			name: "non arn provider to is arn validation",
			validationStruct: struct {
				Test string `validate:"is_arn"`
			}{"bad"},
			errResult: true,
			errString: "failed validation check for 'is_arn', value 'bad' is not a valid arn",
		},
		{
			name: "bad values (-)'s in alpha numeric underscore validation",
			validationStruct: struct {
				Test string `validate:"alphanumunderscore"`
			}{"bad-value"},
			errResult: true,
			errString: "failed validation check for 'alphanumunderscore', on 'Test', value 'bad-value' is invalid",
		},
		{
			name: "bad execute container uri",
			validationStruct: struct {
				Test map[string]string `validate:"valid_execute_container_image"`
			}{map[string]string{
				"execute_container_image_uri": "bad()",
			}},
			errResult: true,
			errString: "failed validation check for 'valid_execute_container_image', value 'map[execute_container_image_uri:bad()]' is an invalid container uri",
		},
		{
			name: "bad pre container uri",
			validationStruct: struct {
				Test map[string]string `validate:"valid_precontainer_image"`
			}{map[string]string{
				"pre_container_image_uri": "bad()",
			}},
			errResult: true,
			errString: "failed validation check for 'valid_precontainer_image', value 'map[pre_container_image_uri:bad()]' is an invalid container uri",
		},
		{
			name: "invalid argument",
			validationStruct: struct {
				Test map[string][]string `validate:"valid_argument"`
			}{map[string][]string{
				"exec": {""}}},
			errResult: true,
			errString: "failed validation check for 'valid_argument', value 'map[exec:[]]' is an invalid argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testValidationStruct := tt.validationStruct
			err := ValidateStruct(testValidationStruct)
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err.Error())
				}
				if err.Error() != tt.errString {
					t.Errorf("\nexpected error '%v', got: '%v'", tt.errString, err.Error())
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateVarErrors(t *testing.T) {
	tests := []struct {
		name        string
		testString  string
		validString string
		errResult   bool
		errString   string
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
			errResult:   true,
			errString:   "failed validation check for 'validate var' 'good'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar("validate var", tt.testString, fmt.Sprintf("eq=%s", tt.validString))
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
				if err.Error() != tt.errString {
					t.Errorf("\nexpected error '%v', got: '%v'", tt.errString, err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error, got: %v", err)
				}
			}
		})
	}
}
