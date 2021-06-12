package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/argoproj-labs/argo-cloudops/cli/internal/api"
	"github.com/spf13/cobra"
)

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:   "get [workflow name]",
	Short: "Gets status of workflow",
	Long:  "Gets status of workflow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		apiCl := api.NewClient(argoCloudOpsServiceAddr())

		status, err := apiCl.GetWorkflowStatus(context.Background(), name)
		if err != nil {
			return err
		}

		output, err := json.Marshal(status)
		if err != nil {
			return fmt.Errorf("unable to generate output, error: %w", err)
		}

		fmt.Println(string(output))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
