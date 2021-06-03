package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	acoEnv "github.com/argoproj-labs/argo-cloudops/internal/env"
)

type CommandVariables struct {
	EnvironmentVariables string
	InitArguments        string
	ExecuteArguments     string
}

type Config struct {
	Version  string
	Commands map[string]map[string]string `yaml:"commands"`
}

func loadConfig() (*Config, error) {
	f, err := ioutil.ReadFile(acoEnv.ConfigFilePath())
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c Config) getCommandDefinition(framework, commandType string) (string, error) {
	if _, ok := c.Commands[framework]; !ok {
		return "", fmt.Errorf("Unknown framework '%s'", framework)
	}

	if _, ok := c.Commands[framework][commandType]; !ok {
		return "", fmt.Errorf("Unknown command type '%s'", commandType)
	}

	return c.Commands[framework][commandType], nil
}

func (c Config) listFrameworks() []string {
	keys := make([]string, len(c.Commands))
	for k := range c.Commands {
		keys = append(keys, k)
	}
	return keys
}

func (c Config) listTypes(framework string) ([]string, error) {
	if _, ok := c.Commands[framework]; !ok {
		return []string{}, fmt.Errorf("Unknown framework '%s'", framework)
	}

	keys := make([]string, 0, len(c.Commands[framework]))
	for k := range c.Commands[framework] {
		keys = append(keys, k)
	}
	return keys, nil
}

func generateExecuteCommand(commandDefinition, environmentVariablesString string, arguments map[string][]string) (string, error) {
	initArguments := ""
	if _, ok := arguments["init"]; ok {
		initArguments = strings.Join(arguments["init"], " ")
	}

	executeArguments := ""
	if _, ok := arguments["execute"]; ok {
		executeArguments = strings.Join(arguments["execute"], " ")
	}

	commandVariables := CommandVariables{
		EnvironmentVariables: environmentVariablesString,
		InitArguments:        initArguments,
		ExecuteArguments:     executeArguments,
	}

	var buf bytes.Buffer
	t, err := template.New("text").Parse(commandDefinition)
	if err != nil {
		return "", err

	}
	err = t.Execute(&buf, commandVariables)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
