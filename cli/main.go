package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/spf13/cobra"
)

// TODO set this and align it to service version
func version() string {
	return "0.0.1"
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func argoCloudOpsServiceAddr() string {
	return getenv("ARGO_CLOUDOPS_SERVICE_ADDR", "https://localhost:8080")
}

func argoCloudOpsUserToken() (string, error) {
	key := "ARGO_CLOUDOPS_USER_TOKEN"
	result := os.Getenv(key)
	if len(result) == 0 {
		return "", errors.New(fmt.Sprintf("%s not found", key))
	}
	return result, nil
}

type createWorkflowRequest struct {
	Arguments            map[string][]string `json:"arguments"`
	EnvironmentVariables map[string]string   `json:"environment_variables"`
	Framework            string              `json:"framework"`
	Parameters           map[string]string   `json:"parameters"`
	ProjectName          string              `json:"project_name"`
	TargetName           string              `json:"target_name"`
	Type                 string              `json:"type"`
	WorkflowTemplateName string              `json:"workflow_template_name"`
}

func newCreateWorkflowRequest(arguments map[string][]string, parameters map[string]string, framework, executeType, environmentVariablesCSV, projectName, targetName, workflowTemplateName string) (*createWorkflowRequest, error) {
	environmentVariables, err := parseEqualsSeparatedCSVToMap(environmentVariablesCSV)
	if err != nil {
		return nil, err
	}

	cr := &createWorkflowRequest{
		Arguments:            arguments,
		EnvironmentVariables: environmentVariables,
		Framework:            framework,
		Parameters:           parameters,
		ProjectName:          projectName,
		TargetName:           targetName,
		Type:                 executeType,
		WorkflowTemplateName: workflowTemplateName,
	}

	return cr, nil
}

type workflowResponse struct {
	WorkflowName string `json:"workflow_name"`
}

type logsResponse struct {
	Logs []string `json:"logs"`
}

func parseEqualsSeparatedCSVToMap(s string) (map[string]string, error) {
	r := make(map[string]string)
	l := strings.Split(s, ",")
	for _, e := range l {
		v := strings.Split(e, "=")
		if len(v) != 2 {
			return r, fmt.Errorf("Could not parse equals separated value %s", e)
		}
		key := v[0]
		value := v[1]
		r[key] = value
	}
	return r, nil
}

func executeWorkflow(cwr *createWorkflowRequest) (string, error) {
	client := &http.Client{}

	requestJSON, err := json.Marshal(cwr)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/workflows", argoCloudOpsServiceAddr())
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", err
	}

	argoCloudOpsUserTkn, err := argoCloudOpsUserToken()
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", argoCloudOpsUserTkn)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("execute workflow request failed with status code %d, message %s", resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var wr workflowResponse

	err = json.Unmarshal(body, &wr)
	if err != nil {
		return "", err
	}

	return wr.WorkflowName, nil
}

func generateArguments(argumentsCSV string) (map[string][]string, error) {
	arguments := make(map[string][]string)

	if argumentsCSV == "" {
		return arguments, nil
	}

	a, err := parseEqualsSeparatedCSVToMap(argumentsCSV)
	if err != nil {
		return arguments, err
	}

	for k, v := range a {
		arguments[k] = strings.Split(v, " ")
	}

	return arguments, nil
}

func generateParameters(parametersCSV string) (map[string]string, error) {
	if parametersCSV != "" {
		parameters, err := parseEqualsSeparatedCSVToMap(parametersCSV)
		if err != nil {
			return make(map[string]string), nil
		}
		return parameters, nil
	}

	return make(map[string]string), nil
}

func printLogStreamOutput(body io.ReadCloser) {
	p := make([]byte, 256)
	for {
		n, err := body.Read(p)
		if err == io.EOF {
			break
		}
		fmt.Print(string(p[:n]))
	}
}

func main() {
	logger := log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), "ts", log.DefaultTimestampUTC)

	if argoCloudOpsServiceAddr() == "https://localhost:8080" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	var argumentsCSV string
	var parametersCSV string
	var framework string
	var environmentVariablesCSV string
	var projectName string
	var targetName string
	var workflowTemplateName string

	var rootCmd = &cobra.Command{
		Use:   "argo-cloudops",
		Short: "Argo CloudOps Command Line Interface",
		Long:  "Argo CloudOps Command Line Interface",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Version Info",
		Long:  "Version Info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version())
		},
	}

	var syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Syncs target with provided arguments",
		Long:  "Syncs target with provided arguments",
		Run: func(cmd *cobra.Command, args []string) {
			arguments, err := generateArguments(argumentsCSV)
			if err != nil {
				level.Error(logger).Log("message", "error generating arguments", "error", err)
				os.Exit(1)
			}

			parameters, err := generateParameters(parametersCSV)
			if err != nil {
				level.Error(logger).Log("message", "error generating workflow parameters", "error", err)
				os.Exit(1)
			}

			cwr, err := newCreateWorkflowRequest(arguments, parameters, framework, "sync", environmentVariablesCSV, projectName, targetName, workflowTemplateName)
			if err != nil {
				level.Error(logger).Log("message", "error creating workflow request", "error", err)
				os.Exit(1)
			}

			result, err := executeWorkflow(cwr)
			if err != nil {
				level.Error(logger).Log("message", "error executing workflow", "error", err)
				os.Exit(1)
			}

			fmt.Printf(result)
		},
	}
	syncCmd.Flags().StringVarP(&argumentsCSV, "arguments", "a", "", "CSV string of equals separated arguments to pass to command (-e Arg1=ValueA,Arg2=ValueB).")
	syncCmd.Flags().StringVarP(&environmentVariablesCSV, "environment_variables", "e", "", "CSV string of equals separated environment variable key value pairs (-e Key1=ValueA,Key2=ValueB)")
	syncCmd.Flags().StringVarP(&framework, "framework", "f", "", "Framework to execute")
	syncCmd.Flags().StringVarP(&parametersCSV, "parameters", "p", "", "CSV string of equals separated parameters name and value (-o Param1=ValueA,Param2=ValueB).")
	syncCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	syncCmd.Flags().StringVarP(&targetName, "target", "t", "", "Name of target")
	syncCmd.Flags().StringVarP(&workflowTemplateName, "workflow_template_name", "w", "", "Name of the workflow template")
	syncCmd.MarkFlagRequired("framework")
	syncCmd.MarkFlagRequired("project_name")
	syncCmd.MarkFlagRequired("target_name")
	syncCmd.MarkFlagRequired("workflow_template_name")

	var diffCmd = &cobra.Command{
		Use:   "diff",
		Short: "Diffs code based on provided arguments",
		Long:  "Diffs code based on provided arguments",
		Run: func(cmd *cobra.Command, args []string) {
			arguments, err := generateArguments(argumentsCSV)
			if err != nil {
				level.Error(logger).Log("message", "error generating arguments", "error", err)
				os.Exit(1)
			}

			parameters, err := generateParameters(parametersCSV)
			if err != nil {
				level.Error(logger).Log("message", "error generating parameters", "error", err)
				os.Exit(1)
			}

			cwr, err := newCreateWorkflowRequest(arguments, parameters, framework, "diff", environmentVariablesCSV, projectName, targetName, workflowTemplateName)
			if err != nil {
				level.Error(logger).Log("message", "error creating workflow request", "error", err)
			}

			result, err := executeWorkflow(cwr)
			if err != nil {
				level.Error(logger).Log("message", "error executing workflow", "error", err)
			}
			fmt.Println(result)
		},
	}
	diffCmd.Flags().StringVarP(&argumentsCSV, "arguments", "a", "", "CSV string of equals separated arguments to pass to command (-e Arg1=ValueA,Arg2=ValueB)")
	diffCmd.Flags().StringVarP(&environmentVariablesCSV, "environment_variables", "e", "", "CSV string of equals separated environment variable key value pairs (-e Key1=ValueA,Key2=ValueB)")
	diffCmd.Flags().StringVarP(&framework, "framework", "f", "", "Framework to execute")
	diffCmd.Flags().StringVarP(&parametersCSV, "parameters", "p", "", "CSV string of equals separated parameters name and value (-o Param1=ValueA,Param2=ValueB).")
	diffCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	diffCmd.Flags().StringVarP(&targetName, "target", "t", "", "Target to apply changes against")
	diffCmd.Flags().StringVarP(&workflowTemplateName, "workflow_template_name", "w", "", "Name of the workflow template")
	diffCmd.MarkFlagRequired("framework")
	diffCmd.MarkFlagRequired("project_name")
	diffCmd.MarkFlagRequired("target_name")
	diffCmd.MarkFlagRequired("workflow_template_name")

	//TODO Validation and better handle all the errors
	var getCmd = &cobra.Command{
		Use:   "get WORKFLOW",
		Short: "Gets status of workflow",
		Long:  "Gets status of workflow",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Printf("Specify workflow name as first argument.")
				os.Exit(1)
			}
			workflowName := args[0]

			client := &http.Client{}

			endpoint := fmt.Sprintf("%s/workflows/%s", argoCloudOpsServiceAddr(), workflowName)
			req, err := http.NewRequest("GET", endpoint, nil)
			if err != nil {
				level.Error(logger).Log("message", "error creating get status request", "error", err)
			}

			//TODO Add authorization once available to logs
			req.Header.Add("Authorization", "NOT_USED_YET")
			resp, err := client.Do(req)
			if err != nil {
				level.Error(logger).Log("message", "error getting status request from service", "error", err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				level.Error(logger).Log("message", "error reading response body from service", "error", err)
			}

			fmt.Println(string(body))
		},
	}

	var streaming bool

	var logsCmd = &cobra.Command{
		Use:   "logs WORKFLOW",
		Short: "Gets logs from a workflow",
		Long:  "Gets logs from a workflow",
		Run: func(cmd *cobra.Command, args []string) {
			client := &http.Client{}

			if len(args) < 1 {
				fmt.Printf("Specify workflow name as first argument.")
				os.Exit(1)
			}

			workflowName := args[0]

			endpoint := fmt.Sprintf("%s/workflows/%s/logs", argoCloudOpsServiceAddr(), workflowName)
			if streaming {
				endpoint = fmt.Sprintf("%s/workflows/%s/logstream", argoCloudOpsServiceAddr(), workflowName)
			}

			req, err := http.NewRequest("GET", endpoint, nil)
			if err != nil {
				level.Error(logger).Log("message", "error creating request for logs from service", "error", err)
			}

			//TODO Add authorization once available to logs
			req.Header.Add("Authorization", "NOT_USED_YET")
			resp, err := client.Do(req)
			if err != nil {
				level.Error(logger).Log("message", "error requesting logs from service", "error", err)
			}

			defer resp.Body.Close()

			if streaming {
				printLogStreamOutput(resp.Body)
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				level.Error(logger).Log("message", "error reading response rom service", "error", err)
			}
			var lr logsResponse
			err = json.Unmarshal(body, &lr)
			if err != nil {
				level.Error(logger).Log("message", "error deserializing logs response rom service", "error", err)
			}
			logs := strings.Join(lr.Logs, "\n")
			fmt.Println(logs)
		},
	}
	logsCmd.Flags().BoolVarP(&streaming, "follow", "f", false, "Follow workflow logs and stream to standard out until workflow is complete")

	var listWorkflowsCmd = &cobra.Command{
		Use:   "list",
		Short: "List workflows for a given project and target",
		Long:  "List workflows for a given project and target",
		Run: func(cmd *cobra.Command, args []string) {
			client := &http.Client{}

			endpoint := fmt.Sprintf("%s/projects/%s/targets/%s/workflows", argoCloudOpsServiceAddr(), projectName, targetName)
			req, err := http.NewRequest("GET", endpoint, nil)
			if err != nil {
				level.Error(logger).Log("message", "error creating list workflows request", "error", err)
			}

			//TODO Add authorization once available to list workflow
			req.Header.Add("Authorization", "NOT_USED_YET")
			resp, err := client.Do(req)
			if err != nil {
				level.Error(logger).Log("message", "error getting request from service", "error", err)
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				level.Error(logger).Log("message", "error reading response body from service", "error", err)
			}

			var workflowIDs []string
			err = json.Unmarshal(body, &workflowIDs)
			if err != nil {
				level.Error(logger).Log("message", "error deserializing response from service", "error", err)
			}

			for _, id := range workflowIDs {
				fmt.Printf("%s\n", id)
			}
		},
	}
	listWorkflowsCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
	listWorkflowsCmd.Flags().StringVarP(&targetName, "target_name", "t", "", "Name of target")
	listWorkflowsCmd.MarkFlagRequired("project_name")
	listWorkflowsCmd.MarkFlagRequired("target_name")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(listWorkflowsCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
