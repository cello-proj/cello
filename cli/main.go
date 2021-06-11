package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/argoproj-labs/argo-cloudops/cli/cmd"
)

// TODO set this and align it to service version
const (
	version = "0.1.1"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func argoCloudOpsServiceAddr() string {
	return getenv("ARGO_CLOUDOPS_SERVICE_ADDR", "https://localhost:8443")
}

func argoCloudOpsUserToken() (string, error) {
	key := "ARGO_CLOUDOPS_USER_TOKEN"
	result := os.Getenv(key)
	if len(result) == 0 {
		return "", fmt.Errorf("%s not found", key)
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

type createGitWorkflowRequest struct {
	Repository string `json:"repository"`
	CommitHash string `json:"sha"`
	Path       string `json:"path"`
	Type       string `json:"type"`
}

func newCreateGitWorkflowRequest(repository, path, sha string) *createGitWorkflowRequest {
	return &createGitWorkflowRequest{
		Repository: repository,
		CommitHash: sha,
		Path:       path,
	}
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

func executeGitWorkflow(cgwr *createGitWorkflowRequest, project, target string) (string, error) {
	client := &http.Client{}

	requestJSON, err := json.Marshal(cgwr)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/projects/%s/targets/%s/operations", argoCloudOpsServiceAddr(), project, target)
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
	// TODO: status should probably be changed to 201?
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return "", fmt.Errorf("execute workflow request failed with status code %d, message %s", resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var wr workflowResponse

	if err = json.Unmarshal(body, &wr); err != nil {
		return "", err
	}

	return wr.WorkflowName, nil
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
	cmd.Execute(version)
}

//func main() {
//	// Set the version for the version command.
//	cmd.Version = version

//	logger := log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), "ts", log.DefaultTimestampUTC)

//	if argoCloudOpsServiceAddr() == "https://localhost:8443" {
//		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
//	}

//	var argumentsCSV string
//	var parametersCSV string
//	var framework string
//	var environmentVariablesCSV string
//	var projectName string
//	var targetName string
//	var workflowTemplateName string
//	var workflowType string
//	var gitRepo string
//	var gitPath string
//	var gitSha string

//	var rootCmd = &cobra.Command{
//		Use:   "argo-cloudops",
//		Short: "Argo CloudOps Command Line Interface",
//		Long:  "Argo CloudOps Command Line Interface",
//		Run: func(cmd *cobra.Command, args []string) {
//			// Do Stuff Here
//		},
//	}

//	// var versionCmd = &cobra.Command{
//	// 	Use:   "version",
//	// 	Short: "Version Info",
//	// 	Long:  "Version Info",
//	// 	Run: func(cmd *cobra.Command, args []string) {
//	// 		fmt.Println(version())
//	// 	},
//	// }

//	var workflowCmd = &cobra.Command{
//		Use:   "workflow",
//		Short: "Creates a workflow with provided arguments",
//		Long:  "Creates a workflow with provided arguments",
//		Run: func(cmd *cobra.Command, args []string) {
//			arguments, err := generateArguments(argumentsCSV)
//			if err != nil {
//				level.Error(logger).Log("message", "error generating arguments", "error", err)
//				os.Exit(1)
//			}

//			parameters, err := generateParameters(parametersCSV)
//			if err != nil {
//				level.Error(logger).Log("message", "error generating workflow parameters", "error", err)
//				os.Exit(1)
//			}

//			cwr, err := newCreateWorkflowRequest(arguments, parameters, framework, workflowType, environmentVariablesCSV, projectName, targetName, workflowTemplateName)
//			if err != nil {
//				level.Error(logger).Log("message", "error creating workflow request", "error", err)
//				os.Exit(1)
//			}

//			result, err := executeWorkflow(cwr)
//			if err != nil {
//				level.Error(logger).Log("message", "error executing workflow", "error", err)
//				os.Exit(1)
//			}

//			fmt.Print(result)
//		},
//	}
//	workflowCmd.Flags().StringVarP(&argumentsCSV, "arguments", "a", "", "CSV string of equals separated arguments to pass to command (-a Arg1=ValueA,Arg2=ValueB).")
//	workflowCmd.Flags().StringVarP(&environmentVariablesCSV, "environment_variables", "e", "", "CSV string of equals separated environment variable key value pairs (-e Key1=ValueA,Key2=ValueB)")
//	workflowCmd.Flags().StringVarP(&framework, "framework", "f", "", "Framework to execute")
//	workflowCmd.Flags().StringVarP(&parametersCSV, "parameters", "p", "", "CSV string of equals separated parameters name and value (-p Param1=ValueA,Param2=ValueB).")
//	workflowCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
//	workflowCmd.Flags().StringVarP(&targetName, "target", "t", "", "Name of target")
//	workflowCmd.Flags().StringVarP(&workflowTemplateName, "workflow_template_name", "w", "", "Name of the workflow template")
//	workflowCmd.Flags().StringVar(&workflowType, "type", "", "Workflow type to execute")
//	workflowCmd.MarkFlagRequired("framework")
//	workflowCmd.MarkFlagRequired("project_name")
//	workflowCmd.MarkFlagRequired("target_name")
//	workflowCmd.MarkFlagRequired("workflow_template_name")
//	workflowCmd.MarkFlagRequired("type")

//	var syncCmd = &cobra.Command{
//		Use:   "sync",
//		Short: "Syncs target using a manifest in git",
//		Long:  "Syncs target using a manifest in git",
//		Run: func(cmd *cobra.Command, args []string) {
//			cgwr := newCreateGitWorkflowRequest(gitRepo, gitPath, gitSha)
//			cgwr.Type = "sync"
//			result, err := executeGitWorkflow(cgwr, projectName, targetName)
//			if err != nil {
//				level.Error(logger).Log("message", "error executing workflow", "error", err)
//				os.Exit(1)
//			}

//			fmt.Print(result)
//		},
//	}
//	syncCmd.Flags().StringVarP(&gitRepo, "repository", "r", "", "Git repository ssh url (e.x. git@github.com:myorg/myrepo.git)")
//	syncCmd.Flags().StringVarP(&gitPath, "path", "p", "", "Path to manifest within git repository")
//	syncCmd.Flags().StringVarP(&gitSha, "sha", "s", "", "Commit sha to use when creating workflow through git")
//	syncCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
//	syncCmd.Flags().StringVarP(&targetName, "target", "t", "", "Name of target")
//	syncCmd.MarkFlagRequired("repository")
//	syncCmd.MarkFlagRequired("path")
//	syncCmd.MarkFlagRequired("sha")
//	syncCmd.MarkFlagRequired("project_name")
//	syncCmd.MarkFlagRequired("target_name")

//	var diffCmd = &cobra.Command{
//		Use:   "diff",
//		Short: "Diff target using a manifest in git",
//		Long:  "Diff target using a manifest in git",
//		Run: func(cmd *cobra.Command, args []string) {
//			cgwr := newCreateGitWorkflowRequest(gitRepo, gitPath, gitSha)
//			cgwr.Type = "diff"
//			result, err := executeGitWorkflow(cgwr, projectName, targetName)
//			if err != nil {
//				level.Error(logger).Log("message", "error executing workflow", "error", err)
//				os.Exit(1)
//			}

//			fmt.Print(result)
//		},
//	}
//	diffCmd.Flags().StringVarP(&gitRepo, "repository", "r", "", "Git repository ssh url (e.x. git@github.com:myorg/myrepo.git)")
//	diffCmd.Flags().StringVarP(&gitPath, "path", "p", "", "Path to manifest within git repository")
//	diffCmd.Flags().StringVarP(&gitSha, "sha", "s", "", "Commit sha to use when creating workflow through git")
//	diffCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
//	diffCmd.Flags().StringVarP(&targetName, "target", "t", "", "Name of target")
//	diffCmd.MarkFlagRequired("repository")
//	diffCmd.MarkFlagRequired("path")
//	diffCmd.MarkFlagRequired("sha")
//	diffCmd.MarkFlagRequired("project_name")
//	diffCmd.MarkFlagRequired("target_name")

//	//TODO Validation and better handle all the errors
//	var getCmd = &cobra.Command{
//		Use:   "get WORKFLOW",
//		Short: "Gets status of workflow",
//		Long:  "Gets status of workflow",
//		Run: func(cmd *cobra.Command, args []string) {
//			if len(args) < 1 {
//				fmt.Printf("Specify workflow name as first argument.")
//				os.Exit(1)
//			}
//			workflowName := args[0]

//			client := &http.Client{}

//			endpoint := fmt.Sprintf("%s/workflows/%s", argoCloudOpsServiceAddr(), workflowName)
//			req, err := http.NewRequest("GET", endpoint, nil)
//			if err != nil {
//				level.Error(logger).Log("message", "error creating get status request", "error", err)
//			}

//			//TODO Add authorization once available to logs
//			req.Header.Add("Authorization", "NOT_USED_YET")
//			resp, err := client.Do(req)
//			if err != nil {
//				level.Error(logger).Log("message", "error getting status request from service", "error", err)
//			}

//			defer resp.Body.Close()
//			body, err := ioutil.ReadAll(resp.Body)
//			if err != nil {
//				level.Error(logger).Log("message", "error reading response body from service", "error", err)
//			}

//			fmt.Println(string(body))
//		},
//	}

//	var streaming bool

//	var logsCmd = &cobra.Command{
//		Use:   "logs WORKFLOW",
//		Short: "Gets logs from a workflow",
//		Long:  "Gets logs from a workflow",
//		Run: func(cmd *cobra.Command, args []string) {
//			client := &http.Client{}

//			if len(args) < 1 {
//				fmt.Printf("Specify workflow name as first argument.")
//				os.Exit(1)
//			}

//			workflowName := args[0]

//			endpoint := fmt.Sprintf("%s/workflows/%s/logs", argoCloudOpsServiceAddr(), workflowName)
//			if streaming {
//				endpoint = fmt.Sprintf("%s/workflows/%s/logstream", argoCloudOpsServiceAddr(), workflowName)
//			}

//			req, err := http.NewRequest("GET", endpoint, nil)
//			if err != nil {
//				level.Error(logger).Log("message", "error creating request for logs from service", "error", err)
//			}

//			//TODO Add authorization once available to logs
//			req.Header.Add("Authorization", "NOT_USED_YET")
//			resp, err := client.Do(req)
//			if err != nil {
//				level.Error(logger).Log("message", "error requesting logs from service", "error", err)
//			}

//			defer resp.Body.Close()

//			if streaming {
//				printLogStreamOutput(resp.Body)
//				return
//			}

//			body, err := ioutil.ReadAll(resp.Body)
//			if err != nil {
//				level.Error(logger).Log("message", "error reading response rom service", "error", err)
//			}
//			var lr logsResponse
//			err = json.Unmarshal(body, &lr)
//			if err != nil {
//				level.Error(logger).Log("message", "error deserializing logs response rom service", "error", err)
//			}
//			logs := strings.Join(lr.Logs, "\n")
//			fmt.Println(logs)
//		},
//	}
//	logsCmd.Flags().BoolVarP(&streaming, "follow", "f", false, "Follow workflow logs and stream to standard out until workflow is complete")

//	var listWorkflowsCmd = &cobra.Command{
//		Use:   "list",
//		Short: "List workflows for a given project and target",
//		Long:  "List workflows for a given project and target",
//		Run: func(cmd *cobra.Command, args []string) {
//			client := &http.Client{}

//			endpoint := fmt.Sprintf("%s/projects/%s/targets/%s/workflows", argoCloudOpsServiceAddr(), projectName, targetName)
//			req, err := http.NewRequest("GET", endpoint, nil)
//			if err != nil {
//				level.Error(logger).Log("message", "error creating list workflows request", "error", err)
//			}

//			//TODO Add authorization once available to list workflow
//			req.Header.Add("Authorization", "NOT_USED_YET")
//			resp, err := client.Do(req)
//			if err != nil {
//				level.Error(logger).Log("message", "error getting request from service", "error", err)
//			}

//			defer resp.Body.Close()
//			body, err := ioutil.ReadAll(resp.Body)
//			if err != nil {
//				level.Error(logger).Log("message", "error reading response body from service", "error", err)
//			}

//			var workflowIDs []string
//			err = json.Unmarshal(body, &workflowIDs)
//			if err != nil {
//				level.Error(logger).Log("message", "error deserializing response from service", "error", err)
//			}

//			for _, id := range workflowIDs {
//				fmt.Printf("%s\n", id)
//			}
//		},
//	}
//	listWorkflowsCmd.Flags().StringVarP(&projectName, "project_name", "n", "", "Name of project")
//	listWorkflowsCmd.Flags().StringVarP(&targetName, "target_name", "t", "", "Name of target")
//	listWorkflowsCmd.MarkFlagRequired("project_name")
//	listWorkflowsCmd.MarkFlagRequired("target_name")

//	rootCmd.AddCommand(cmd.versionCmd)
//	rootCmd.AddCommand(workflowCmd)
//	rootCmd.AddCommand(syncCmd)
//	rootCmd.AddCommand(diffCmd)
//	rootCmd.AddCommand(getCmd)
//	rootCmd.AddCommand(logsCmd)
//	rootCmd.AddCommand(listWorkflowsCmd)

//	if err := rootCmd.Execute(); err != nil {
//		os.Exit(1)
//	}
// }
