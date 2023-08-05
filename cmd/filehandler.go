// filehandler.go
package cmd

import (
	"fmt"
	"os"
)

func ValidateFile(filePath string) error {
	// Check if the file exists and is a regular file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}
		return fmt.Errorf("error checking file: %s", err.Error())
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("provided path is a directory, not a file")
	}

	return nil
}
