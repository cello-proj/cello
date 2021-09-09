//go:build !test
// +build !test

// Package helpers should be temporary until the cli args/flags are refactored.
// TODO remove this package.
package helpers

import (
	"fmt"
	"strings"
)

// GenerateParameters converts a csv of equals separated values into a map of strings.
func GenerateParameters(parametersCSV string) (map[string]string, error) {
	if parametersCSV != "" {
		parameters, err := ParseEqualsSeparatedCSVToMap(parametersCSV)
		if err != nil {
			return make(map[string]string), err
		}
		return parameters, nil
	}

	return make(map[string]string), nil
}

// GenerateArguments converts a csv of equals separate values into a map of string slices.
func GenerateArguments(argumentsCSV string) (map[string][]string, error) {
	arguments := make(map[string][]string)

	if argumentsCSV == "" {
		return arguments, nil
	}

	a, err := ParseEqualsSeparatedCSVToMap(argumentsCSV)
	if err != nil {
		return arguments, err
	}

	for k, v := range a {
		arguments[k] = strings.Split(v, " ")
	}

	return arguments, nil
}

// ParseEqualsSeparatedCSVToMap converts a csv of equals separated values into a map of strings.
func ParseEqualsSeparatedCSVToMap(s string) (map[string]string, error) {
	r := make(map[string]string)
	l := strings.Split(s, ",")
	for _, e := range l {
		v := strings.SplitN(e, "=", 2)
		if len(v) != 2 {
			return r, fmt.Errorf("could not parse equals separated value %s", e)
		}
		key := v[0]
		value := v[1]
		r[key] = value
	}
	return r, nil
}
