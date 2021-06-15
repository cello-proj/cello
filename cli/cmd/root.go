package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "argo-cloudops",
	Short: "Argo CloudOps Command Line Interface",
	Long:  "Argo CloudOps Command Line Interface",
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
	repository              string
	sha                     string
	streamLogs              bool
	targetName              string
	workflowTemplateName    string
	workflowType            string

	// TODO populate during build/release
	version = "0.1.1"
)

// Execute adds all child commands to the root command and sets flags
// appropriately.  This is called by main.main(). It only needs to happen once
// to the rootCmd.
func Execute(versionInfo string) {
	version = versionInfo
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.foo.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func argoCloudOpsServiceAddr() string {
	addr := os.Getenv("ARGO_CLOUDOPS_SERVICE_ADDR")
	if addr == "" {
		addr = "https://localhost:8443"
	}
	return addr
}

func argoCloudOpsUserToken() (string, error) {
	key := "ARGO_CLOUDOPS_USER_TOKEN"
	result := os.Getenv(key)
	if len(result) == 0 {
		return "", fmt.Errorf("%s not found", key)
	}
	return result, nil
}
