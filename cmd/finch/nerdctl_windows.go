package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func convertToWSLPath(winPath string) (string, error) {
	var path = filepath.Clean(winPath)
	var err error
	if !filepath.IsAbs(winPath) {
		path, err = filepath.Abs(winPath)
		if err != nil {
			return "", err
		}
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

// Copied from https://github.com/rancher-sandbox/rancher-desktop/blob/5cfeb80aba3f9c9d85d840ed1143caed06d21c02/src/go/nerdctl-stub/main_windows.go#L69
func handleVolume(v string) (string, error) {
	cleanArg := v
	readWrite := ""
	if strings.HasSuffix(v, ":ro") || strings.HasSuffix(v, ":rw") {
		readWrite = v[len(v)-3:]
		cleanArg = v[:len(v)-3]
	}
	// For now, assume the container path doesn't contain colons.
	colonIndex := strings.LastIndex(cleanArg, ":")
	if colonIndex < 0 {
		return "", fmt.Errorf("invalid volume mount: %s does not contain : separator", v)
	}
	hostPath := cleanArg[:colonIndex]
	containerPath := cleanArg[colonIndex+1:]
	wslHostPath, err := convertToWSLPath(hostPath)
	if err != nil {
		return "", fmt.Errorf("could not get volume host path for %s: %w", v, err)
	}
	return wslHostPath + ":" + containerPath + readWrite, nil
}
