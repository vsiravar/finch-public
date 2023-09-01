//go:build windows

package main

func (nc *nerdctlCommand) GetLimaArgs() []string {
	wd, err := nc.systemDeps.GetWd()
	if err != nil {
		nc.logger.Warnln("failed to get working directory, will default to user home")
		return append([]string{"shell", limaInstanceName, "sudo", "-E"})
	}
	wslPath := convertToWSLPath(wd)
	return append([]string{"shell", "--workdir", wslPath, limaInstanceName, "sudo", "-E"})
}
