package util

import (
	"bufio"
	"os"

	"github.com/latam-airlines/crane/logger"
)

func FileExists(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}

func ParseMultiFileLinesToArray(envFiles []string) ([]string, error) {
	var envs []string

	for _, path := range envFiles {
		parsedEnvs, err := ParseSingleFileLinesToArray(path)
		if err != nil {
			return nil, err
		}
		envs = append(envs, parsedEnvs...)
	}

	return envs, nil
}

func ParseSingleFileLinesToArray(path string) ([]string, error) {
	logger.Instance().Debugf("Parseando el archivo %s", path)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	logger.Instance().Debugln("Parseo exitoso")
	return lines, scanner.Err()
}
