package utils

import (
	"fmt"
	"strings"
)

const urlVersion = "v2"

func StringSlice2Map(slice []string) map[string]string {
	envs := make(map[string]string)
	for _, val := range slice {
		splitted := strings.Split(val, "=")
		if len(splitted) != 2 {
			continue
		}
		envs[splitted[0]] = splitted[1]
	}
	return envs
}

func ValidateEndpoint(endpoint string) string {
	if strings.Contains(endpoint, urlVersion) {
		return endpoint
	} else {
		return endpoint + "/" + urlVersion
	}
}

func ExtractString(params map[string]interface{}, key string) string {
	val := ""
	valI, ok := params[key]
	if ok {
		val = fmt.Sprint(valI)
	}
	return val
}
