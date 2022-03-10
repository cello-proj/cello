//go:build !test
// +build !test

package cmd

import (
	"context"
	"fmt"

	"github.com/cello-proj/cello/cli/internal/api"

	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Executes an operation on a project target using a manifest in git",
	Long:  "Executes an operation on a project target using a manifest in git",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := argoCloudOpsUserToken()
		if err != nil {
			cobra.CheckErr(err)
		}

		apiCl := api.NewClient(argoCloudOpsServiceAddr(), token)

		resp, err := apiCl.Exec(context.Background(), api.TargetOperationInput{Path: gitPath, ProjectName: projectName, SHA: gitSHA, TargetName: targetName})
		if err != nil {
			cobra.CheckErr(err)
		}

		// Our current contract is to output only the name.
		fmt.Print(resp.WorkflowName)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)

	// TODO these should be '-' separated.
	execCmd.Flags().StringVarP(&gitPath, "path", "p", "", "Path to manifest within git repository")
	execCmd.Flags().StringVarP(&gitSHA, "sha", "s", "", "Commit sha to use when creating workflow through git")
	execCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	// TODO inconsistent
	execCmd.Flags().StringVarP(&targetName, "target", "t", "", "Name of target")

	execCmd.MarkFlagRequired("path")
	execCmd.MarkFlagRequired("sha")
	execCmd.MarkFlagRequired("project_name")
	execCmd.MarkFlagRequired("target_name")
}
