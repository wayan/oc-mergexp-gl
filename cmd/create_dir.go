package cmd

import (
	"fmt"
	"os"
)

func createDirIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dirPath, 0755)
			if err != nil {
				return fmt.Errorf("Error creating directory: %w\n", err)
			}
			return nil
		}
		// Some other error occurred when trying to stat the directory
		return fmt.Errorf("Error checking directory existence: %w\n", err)
	}
	return nil
}
