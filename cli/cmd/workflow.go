//go:build !test
// +build !test

package cmd

import (
	"context"
	"fmt"

	"github.com/cello-proj/cello/cli/internal/api"
	"github.com/cello-proj/cello/cli/internal/helpers"
	"github.com/cello-proj/cello/internal/requests"

	"github.com/spf13/cobra"
)

// workflowCmd represents the workflow command
// TODO should this really be 'exec'?
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Creates a workflow execution with provided arguments",
	Long:  "Creates a workflow execution with provided arguments",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := argoCloudOpsUserToken()
		if err != nil {
			cobra.CheckErr(err)
		}

		// TODO this should be removed in favor of supporting multiple flags.
		arguments, err := helpers.GenerateArguments(argumentsCSV)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("unable to generate arguments, error: %w", err))
		}

		// TODO this should be removed in favor of supporting multiple flags.
		envVars, err := helpers.ParseEqualsSeparatedCSVToMap(environmentVariablesCSV)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("unable to generate parameters, error: %w", err))
		}

		// TOOD this should be removed in favor of supporting multiple flags.
		parameters, err := helpers.GenerateParameters(parametersCSV)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("unable to generate parameters, error: %w", err))
		}

		apiCl := api.NewClient(argoCloudOpsServiceAddr(), token)

		input := requests.CreateWorkflow{
			Arguments:            arguments,
			EnvironmentVariables: envVars,
			Framework:            framework,
			Parameters:           parameters,
			ProjectName:          projectName,
			TargetName:           targetName,
			Type:                 workflowType,
			WorkflowTemplateName: workflowTemplateName,
		}

		resp, err := apiCl.ExecuteWorkflow(context.Background(), input)
		if err != nil {
			cobra.CheckErr(err)
		}

		// Our current contract is to output only the name.
		fmt.Print(resp.WorkflowName)
	},
}

func init() {
	rootCmd.AddCommand(workflowCmd)

	// TODO these should be `-` separated.

	// TODO we should accept multiple flags instead of csv
	workflowCmd.Flags().StringVarP(&argumentsCSV, "arguments", "a", "", "CSV string of equals separated arguments to pass to command (-a Arg1=ValueA,Arg2=ValueB).")

	// TODO we should accept multiple flags instead of csv
	workflowCmd.Flags().StringVarP(&environmentVariablesCSV, "environment_variables", "e", "", "CSV string of equals separated environment variable key value pairs (-e Key1=ValueA,Key2=ValueB)")
	workflowCmd.Flags().StringVarP(&framework, "framework", "f", "", "Framework to execute")

	// TODO we should accept multiple flags instead of csv
	workflowCmd.Flags().StringVarP(&parametersCSV, "parameters", "p", "", "CSV string of equals separated parameters name and value (-p Param1=ValueA,Param2=ValueB).")
	// TODO this and target aren't consistent
	workflowCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	workflowCmd.Flags().StringVarP(&targetName, "target", "t", "", "Name of target")
	workflowCmd.Flags().StringVarP(&workflowTemplateName, "workflow_template_name", "w", "", "Name of the workflow template")
	workflowCmd.Flags().StringVar(&workflowType, "type", "", "Workflow type to execute")

	workflowCmd.MarkFlagRequired("framework")
	workflowCmd.MarkFlagRequired("project_name")
	workflowCmd.MarkFlagRequired("target_name")
	workflowCmd.MarkFlagRequired("workflow_template_name")
	workflowCmd.MarkFlagRequired("type")
}
