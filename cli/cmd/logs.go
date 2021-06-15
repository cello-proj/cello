package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/argoproj-labs/argo-cloudops/cli/internal/api"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [workflow name]",
	Short: "Gets logs from a workflow",
	Long:  "Gets logs from a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowName := args[0]

		apiCl := api.NewClient(argoCloudOpsServiceAddr(), "")

		if streamLogs {
			// This is a _very_ simple approach to streaming.
			return apiCl.StreamLogs(context.Background(), os.Stdout, workflowName)
		}

		resp, err := apiCl.GetLogs(context.Background(), workflowName)
		if err != nil {
			return err
		}

		fmt.Println(strings.Join(resp.Logs, "\n"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolVarP(&streamLogs, "follow", "f", false, "Follow workflow logs and stream to standard out until workflow is complete")
}
