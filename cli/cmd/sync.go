//go:build !test
// +build !test

package cmd

import (
	"context"
	"fmt"

	"github.com/cello-proj/cello/cli/internal/api"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs a project target using a manifest in git",
	Long:  "Syncs a project target using a manifest in git",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := argoCloudOpsUserToken()
		if err != nil {
			cobra.CheckErr(err)
		}

		apiCl := api.NewClient(argoCloudOpsServiceAddr(), token)

		resp, err := apiCl.Sync(context.Background(), api.TargetOperationInput{Path: gitPath, ProjectName: projectName, SHA: gitSHA, TargetName: targetName})
		if err != nil {
			cobra.CheckErr(err)
		}

		// Our current contract is to output only the name.
		fmt.Print(resp.WorkflowName)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// TODO these should be '-' separated.
	syncCmd.Flags().StringVarP(&gitPath, "path", "p", "", "Path to manifest within git repository")
	syncCmd.Flags().StringVarP(&gitSHA, "sha", "s", "", "Commit sha to use when creating workflow through git")
	syncCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	// TODO inconsistent
	syncCmd.Flags().StringVarP(&targetName, "target", "t", "", "Name of target")

	syncCmd.MarkFlagRequired("path")
	syncCmd.MarkFlagRequired("sha")
	syncCmd.MarkFlagRequired("project_name")
	syncCmd.MarkFlagRequired("target_name")
}
