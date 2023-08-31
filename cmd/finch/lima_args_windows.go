//go:build windows

package main

import (
	"path/filepath"
	"strings"
)

func (nc *nerdctlCommand) GetLimaArgs() []string {
	wd, err := nc.systemDeps.GetWd()
	if err != nil {
		nc.logger.Warnln("failed to get working directory, will default to user home")
		return append([]string{"shell", limaInstanceName, "sudo", "-E"})
	}
	wslPath := convertToWSLPath(wd)
	return append([]string{"shell", "--workdir", wslPath, limaInstanceName, "sudo", "-E"})
}

func convertToWSLPath(winPath string) string {
	path := filepath.Clean(winPath)

	if len(path) >= 2 && path[1] == ':' {
		drive := strings.ToLower(string(path[0]))
		remainingPath := ""
		if len(path) > 3 {
			remainingPath = path[3:]
		}
		return filepath.ToSlash(filepath.Join(string(filepath.Separator), "mnt", drive, remainingPath))
	}
	return path
}
