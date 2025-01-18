package util

import "strings"

func OptionsToMap(options string) map[string]string {
	opts := map[string]string{}
	options = strings.TrimSpace(options)

	if options != "" {
		kvPairs := strings.Split(options, " ")
		for _, entry := range kvPairs {
			kv := strings.SplitN(entry, "=", 2)
			opts[kv[0]] = kv[1]
		}
	}

	return opts
}
