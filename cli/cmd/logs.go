package cmd

import (
	"context"
	"fmt"
	"io"
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
			r, w := io.Pipe()

			// TODO review this code
			// Spawn a reader ahead of time.
			go func() {
				b := make([]byte, 256)
				for {
					n, err := r.Read(b)
					if err == io.EOF {
						break
					}
					fmt.Print(string(b[:n]))
				}
			}()

			if err := apiCl.StreamLogs(context.Background(), w, workflowName); err != nil {
				return err
			}

			return nil
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
