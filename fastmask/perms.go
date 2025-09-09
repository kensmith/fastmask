package main

import (
	"fmt"
	"os"
)

func checkDirectoryPermissions(dirname string) error {
	dirInfo, err := os.Stat(dirname)
	if err != nil {
		return fmt.Errorf("cannot access config directory: %w", err)
	}
	dirMode := dirInfo.Mode().Perm()
	if dirMode != _configDirPerm {
		return fmt.Errorf("insecure config directory permissions %04o (should be %04o): %s", dirMode, _configDirPerm, dirname)
	}

	return nil
}

func checkFilePermissions(filename string) error {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("cannot access config file: %w", err)
	}
	fileMode := fileInfo.Mode().Perm()
	if fileMode != _configFilePerm && fileMode != _configFilePermReadOnly {
		return fmt.Errorf("insecure config file permissions %04o (should be %04o or %04o): %s", fileMode, _configFilePerm, _configFilePermReadOnly, filename)
	}

	return nil
}
