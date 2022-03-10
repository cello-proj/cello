//go:build !test
// +build !test

package cmd

import (
	"context"
	"fmt"

	"github.com/cello-proj/cello/cli/internal/api"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflow executions for a given project and target",
	Long:  "List workflow executions for a given project and target",
	Run: func(cmd *cobra.Command, args []string) {
		apiCl := api.NewClient(argoCloudOpsServiceAddr(), "")

		resp, err := apiCl.GetWorkflows(context.Background(), projectName, targetName)
		if err != nil {
			cobra.CheckErr(err)
		}

		for _, w := range resp {
			fmt.Printf("%s\n", w)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// TODO this is our current contract. These should really be `-` separated.
	listCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	listCmd.Flags().StringVarP(&targetName, "target_name", "t", "", "Name of target")

	listCmd.MarkFlagRequired("project_name")
	listCmd.MarkFlagRequired("target_name")
}
