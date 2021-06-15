// +build !test

package cmd

import (
	"context"
	"fmt"

	"github.com/argoproj-labs/argo-cloudops/cli/internal/api"
	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff a project target using a manifest in git",
	Long:  "Diff a project target using a manifest in git",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := argoCloudOpsUserToken()
		if err != nil {
			return err
		}

		apiCl := api.NewClient(argoCloudOpsServiceAddr(), token)

		resp, err := apiCl.Diff(context.Background(), projectName, targetName, gitSHA, gitPath)
		if err != nil {
			return err
		}

		// Our current contract is to output only the name.
		fmt.Print(resp.WorkflowName)

		return nil
	},
}

func init() {
	// TODO these should be '-' separated.
	diffCmd.Flags().StringVarP(&gitPath, "path", "p", "", "Path to manifest within git repository")
	diffCmd.Flags().StringVarP(&gitSHA, "sha", "s", "", "Commit sha to use when creating workflow through git")
	diffCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	// TODO is this correct (inconsistent)?
	diffCmd.Flags().StringVarP(&targetName, "target", "t", "", "Name of target")

	diffCmd.MarkFlagRequired("repository")
	diffCmd.MarkFlagRequired("path")
	diffCmd.MarkFlagRequired("sha")
	diffCmd.MarkFlagRequired("project_name")
	diffCmd.MarkFlagRequired("target_name")
}
