package cmd

import "os"

// returns path of default deployment key
func findDefaultDeployKey() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	path := homeDir + "/.ssh/absbot_rsa"
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}
