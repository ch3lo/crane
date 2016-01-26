package utils

import (
	"strings"
)

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
