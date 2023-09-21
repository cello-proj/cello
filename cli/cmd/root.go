package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cello",
	Short: "Cello Command Line Interface",
	Long:  "Cello Command Line Interface",
}

var (
	// Flags
	argumentsCSV            string
	environmentVariablesCSV string
	framework               string
	gitPath                 string
	gitSHA                  string
	parametersCSV           string
	projectName             string
	streamLogs              bool
	targetName              string
	workflowTemplateName    string
	workflowType            string

	// This is set here so we can access it in this package.
	version string
)

// Execute adds all child commands to the root command and sets flags
// appropriately.  This is called by main.main(). It only needs to happen once
// to the rootCmd.
func Execute(versionInfo string) {
	version = versionInfo
	cobra.CheckErr(rootCmd.Execute())
}

// For root level flags
func init() {
}

// TODO refactor
func argoCloudOpsServiceAddr() string {
	legacyKey := "ARGO_CLOUDOPS_SERVICE_ADDR"
	key := "CELLO_SERVICE_ADDR"
	addr := envOrLegacy(key, legacyKey)
	if addr == "" {
		addr = "https://localhost:8443"
	}
	return addr
}

// TODO refactor
func argoCloudOpsUserToken() (string, error) {
	legacyKey := "ARGO_CLOUDOPS_USER_TOKEN" // #nosec G101
	key := "CELLO_USER_TOKEN"
	result := envOrLegacy(key, legacyKey)
	if len(result) == 0 {
		return "", fmt.Errorf("%s not found", key)
	}
	return result, nil
}

func envOrLegacy(key, legacyKey string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return os.Getenv(legacyKey)
}
