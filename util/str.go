package util

import (
	"strings"
	"unicode"
	)

func MaskEnv(unmaskedEnvs []string) []string {
	var maskedEnvs []string
	for _, val := range unmaskedEnvs {
		kv := strings.Split(val, "=")
		if strings.Contains(kv[0], "pass") {
			maskedEnvs = append(maskedEnvs, kv[0]+"="+"*****")
		} else {
			maskedEnvs = append(maskedEnvs, val)
		}
	}

	return maskedEnvs
}

var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func Letter(n int) string {
	return string(letters[n])
}

// StringSlice2Map translates [deep=purple jimmy=hendrix] into map[deep:purple jimmy:hendrix]
func StringSlice2Map(slice []string) map[string]string {
	themap := make(map[string]string)
	for _, keyvalue := range slice {
		f := func(c rune) bool { return unicode.IsSymbol(c) }
		r := strings.FieldsFunc(keyvalue, f)
		themap[r[0]]=r[1]
	}
	return themap
}
