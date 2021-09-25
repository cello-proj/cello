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
	Run: func(cmd *cobra.Command, args []string) {
		workflowName := args[0]

		apiCl := api.NewClient(argoCloudOpsServiceAddr(), "")

		ctx := context.Background()
		if streamLogs {
			// This is a _very_ simple approach to streaming.
			err := apiCl.StreamLogs(ctx, os.Stdout, workflowName)
			// catch and retry stream internal error
			for err != nil && strings.Contains(err.Error(), "stream error: stream ID 1; INTERNAL_ERROR") {
				err = apiCl.StreamLogs(ctx, os.Stdout, workflowName)
			}
			cobra.CheckErr(err)

		} else {
			resp, err := apiCl.GetLogs(ctx, workflowName)
			if err != nil {
				cobra.CheckErr(err)
			}
			fmt.Println(strings.Join(resp.Logs, "\n"))
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolVarP(&streamLogs, "follow", "f", false, "Follow workflow logs and stream to standard out until workflow is complete")
}
