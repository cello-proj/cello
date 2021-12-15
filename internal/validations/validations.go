package validations

import (
	"path/filepath"
	"regexp"

	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/distribution/distribution/reference"
)

var (
	imageURIs []string
)

// SetImageURIs restricts the approved container URIs to the provided set. To reset to a default allow-all state,
// provide an empty list.
func SetImageURIs(uris []string) {
	imageURIs = uris
}

// Validate iterates through the provided validation funcs.
func Validate(validations ...func() error) error {
	for _, v := range validations {
		if err := v(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateStruct validates the provided struct.
func ValidateStruct(input interface{}) error {
	customValidators := map[string]govalidator.CustomTypeValidator{
		"alphanumunderscore": isAlphaNumbericUnderscore,
	}

	for k, v := range customValidators {
		if _, exists := govalidator.CustomTypeTagMap.Get(k); !exists {
			govalidator.CustomTypeTagMap.Set(k, v)
		}
	}

	_, err := govalidator.ValidateStruct(input)
	return err
}

// isAlphaNumbericUnderscore
func isAlphaNumbericUnderscore(field interface{}, kind interface{}) bool {
	// only handle strings
	switch s := field.(type) {
	case string:
		// Vault does not allow dashes and must start with alpha.
		pattern := `^([a-zA-Z])[a-zA-Z0-9_]*$`
		return regexp.MustCompile(pattern).MatchString(s)
	default:
		panic("unsupported field type for isAlphaNumbericUnderscore2")
	}
}

// IsValidARN determines if the string is a valid AWS ARN.
func IsValidARN(s string) bool {
	return arn.IsARN(s)
}

// IsValidImageURI determines if the image URI is a valid container image URI
// format.
func IsValidImageURI(imageURI string) bool {
	_, err := reference.ParseAnyReference(imageURI)
	return err == nil
}

// IsApprovedImageURI determines if the image URI is approved for use. Default is allow-all. Full filepath matching
// rules are applied to allow varying levels of globbing and wildcards.
// Examples:
// - Direct match: docker.myco.com/argocloups/cdk:1234
// - Direct image, any tag: docker.myco.com/argocloudops/cdk:*
// - Any image within a specific registry: docker.myco.com/*/*
func IsApprovedImageURI(imageURI string) bool {
	// default to allow all if no restrictions set
	if len(imageURIs) == 0 {
		return true
	}

	for _, pattern := range imageURIs {
		if ok, _ := filepath.Match(pattern, imageURI); ok {
			return true
		}
	}

	return false
}

// IsValidGitURI determines if the provided string is a valid git URI.
func IsValidGitURI(s string) bool {
	pattern := `((git|ssh|https)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`
	return regexp.MustCompile(pattern).MatchString(s)
}
