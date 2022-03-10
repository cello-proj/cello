package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cello-proj/cello/cli/internal/api"

	"github.com/spf13/cobra"
)

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:   "get [workflow name]",
	Short: "Gets status of workflow",
	Long:  "Gets status of workflow",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		apiCl := api.NewClient(argoCloudOpsServiceAddr(), "")

		status, err := apiCl.GetWorkflowStatus(context.Background(), name)
		if err != nil {
			cobra.CheckErr(err)
		}

		// Our current "contract" is to output json.
		output, err := json.Marshal(status)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("unable to generate output, error: %w", err))
		}

		fmt.Println(string(output))
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
