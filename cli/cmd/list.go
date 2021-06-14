// +build !test

package cmd

import (
	"context"
	"fmt"

	"github.com/argoproj-labs/argo-cloudops/cli/internal/api"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflow executions for a given project and target",
	Long:  "List workflow executions for a given project and target",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiCl := api.NewClient(argoCloudOpsServiceAddr())

		resp, err := apiCl.GetWorkflows(context.Background(), projectName, targetName)
		if err != nil {
			return err
		}

		for _, w := range resp {
			fmt.Printf("%s\n", w)
		}

		return nil
	},
}

func init() {
	// TODO this is our current contract. These should really be `-` separated.
	listCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	listCmd.Flags().StringVarP(&targetName, "target_name", "t", "", "Name of target")
	listCmd.MarkFlagRequired("project_name")
	listCmd.MarkFlagRequired("target_name")

	rootCmd.AddCommand(listCmd)
}
