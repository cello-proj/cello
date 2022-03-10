package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cello-proj/cello/cli/internal/api"

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
			cobra.CheckErr(apiCl.StreamLogs(ctx, os.Stdout, workflowName))
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
