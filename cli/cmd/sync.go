// +build !test

package cmd

import (
	"context"
	"fmt"

	"github.com/argoproj-labs/argo-cloudops/cli/internal/api"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs a project target using a manifest in git",
	Long:  "Syncs a project target using a manifest in git",
	RunE: func(cmd *cobra.Command, args []string) error {

		token, err := argoCloudOpsUserToken()
		if err != nil {
			return err
		}

		apiCl := api.NewClient(argoCloudOpsServiceAddr(), token)

		resp, err := apiCl.Sync(context.Background(), projectName, targetName, gitSHA, gitPath)
		if err != nil {
			return err
		}

		// Our current contract is to output only the name.
		fmt.Print(resp.WorkflowName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVarP(&gitPath, "path", "p", "", "Path to manifest within git repository")
	syncCmd.Flags().StringVarP(&gitSHA, "sha", "s", "", "Commit sha to use when creating workflow through git")
	syncCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	// TODO is this correct (inconsistent)?
	syncCmd.Flags().StringVarP(&targetName, "target", "t", "", "Name of target")

	syncCmd.MarkFlagRequired("path")
	syncCmd.MarkFlagRequired("sha")
	syncCmd.MarkFlagRequired("project_name")
	syncCmd.MarkFlagRequired("target_name")
}