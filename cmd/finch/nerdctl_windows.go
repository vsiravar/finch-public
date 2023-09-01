package main

import (
	"path/filepath"
	"strings"
)

func convertToWSLPath(winPath string) (string, error) {
	var path string
	var err error
	if !filepath.IsAbs(winPath) {
		path, err = filepath.Abs(winPath)
		return "", err
	}
	if len(path) >= 2 && path[1] == ':' {
		drive := strings.ToLower(string(path[0]))
		remainingPath := ""
		if len(path) > 3 {
			remainingPath = path[3:]
		}
		return filepath.ToSlash(filepath.Join(string(filepath.Separator), "mnt", drive, remainingPath)), nil
	}
	return path, nil
}

func handleFilePath(filePath string) (string, error) {
	return convertToWSLPath(filePath)
}
