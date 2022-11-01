package validations

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	fooErr := errors.New("foo")
	tests := []struct {
		name        string
		validations []func() error
		wantErr     error
	}{
		{
			name: "all passing",
			validations: []func() error{
				func() error { return nil },
				func() error { return nil },
			},
		},
		{
			name: "one fails",
			validations: []func() error{
				func() error { return nil },
				func() error { return fooErr },
			},
			wantErr: fooErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, Validate(tt.validations...))
		})
	}
}

func TestIsAlphaNumericUnderscore(t *testing.T) {
	type testStruct struct {
		Test string `valid:"alphanumunderscore"`
	}

	tests := []struct {
		name       string
		testString string
		wantErr    error
	}{
		{
			name:       "valid alpha numeric underscore",
			testString: "abcd1234____",
		},
		{
			name:       "invalid alpha numeric underscore characters",
			testString: "--[[]]  ",
			wantErr:    fmt.Errorf("Test: --[[]]   does not validate as alphanumunderscore"),
		},
		{
			name:       "doesn't start with alpha",
			testString: "0asdlkfj",
			wantErr:    fmt.Errorf("Test: 0asdlkfj does not validate as alphanumunderscore"),
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

func TestIsValidGitURI(t *testing.T) {
	tests := []struct {
		name       string
		testString string
		want       bool
	}{
		{
			name:       "valid https",
			testString: "https://github.com/cello-proj/cello.git",
			want:       true,
		},
		{
			name:       "valid git",
			testString: "git@github.com:cello-proj/cello.git",
			want:       true,
		},
		{
			name:       "valid ssh",
			testString: "ssh://bob@example.com:22/path/to/repo.git/",
			want:       true,
		},
		{
			name:       "invalid shorthand",
			testString: "cello-proj/cello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidGitURI(tt.testString))
		})
	}
}

func TestIsValidImageURI(t *testing.T) {
	tests := []struct {
		name       string
		testString string
		want       bool
	}{
		{
			name:       "valid execute container image",
			testString: "celloproj/cello-cdk:1.87.1",
			want:       true,
		},
		{
			name:       "invalid execute container image",
			testString: "()cello  -- /cello-cdk:1.87.1",
		},
		{
			name: "no image provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidImageURI(tt.testString))
		})
	}
}

func TestIsApprovedImageURI(t *testing.T) {
	tests := []struct {
		name       string
		testString string
		want       bool
		uris       []string
	}{
		{
			name:       "default allow-all passes",
			testString: "celloproj/cello-cdk:1.87.1",
			want:       true,
		},
		{
			name:       "default allow-all passes with multi-separator",
			testString: "docker.myco.com/slash1/celloproj/slash2/cello-cdk:latest",
			want:       true,
			uris:       []string{},
		},
		{
			name:       "direct matched image from config passes",
			testString: "cello/match:1.87.1",
			want:       true,
			uris:       []string{"cello/match:1.87.1"},
		},
		{
			name:       "rejects non-approved image",
			testString: "cello/nomatch:1.87.1",
			want:       false,
			uris:       []string{"cello/match:1.87.1"},
		},
		{
			name:       "matches globbing on tag",
			testString: "cello/match:1.87.1",
			want:       true,
			uris:       []string{"cello/match:*"},
		},
		{
			name:       "matches globbing on any image within a registry",
			testString: "docker.myco.com/cello/match:1.87.1",
			want:       true,
			uris:       []string{"docker.myco.com/*/*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetImageURIs(tt.uris)

			assert.Equal(t, tt.want, IsApprovedImageURI(tt.testString))
		})
	}
}

func TestIsValidARN(t *testing.T) {
	tests := []struct {
		name       string
		testString string
		want       bool
	}{
		{
			name:       "valid arn",
			testString: "arn:aws:iam::012345678901:policy/test-policy",
			want:       true,
		},
		{
			name:       "invalid arn",
			testString: "invalid-arn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidARN(tt.testString))
		})
	}
}
