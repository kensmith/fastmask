package main

import (
	"fmt"
	"os"
	"runtime"
)

func shouldCheckPermissions() bool {
	supportsFSPerms := map[string]bool{
		"linux":   true,
		"darwin":  true, // macOS
		"freebsd": true,
		"openbsd": true,
		"netbsd":  true,
		"unix":    true,
	}

	return supportsFSPerms[runtime.GOOS]
}

func CheckDirectoryPermissions(dirname string) error {
	dirInfo, err := os.Stat(dirname)
	if err != nil {
		return fmt.Errorf("cannot access config directory: %w", err)
	}

	if shouldCheckPermissions() {
		dirMode := dirInfo.Mode().Perm()
		if dirMode != _configDirPerm {
			return fmt.Errorf("insecure config directory permissions %04o (should be %04o): %s", dirMode, _configDirPerm, dirname)
		}
	}

	return nil
}

func CheckFilePermissions(filename string) error {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("cannot access config file: %w", err)
	}

	if shouldCheckPermissions() {
		fileMode := fileInfo.Mode().Perm()
		if fileMode != _configFilePerm && fileMode != _configFilePermReadOnly {
			return fmt.Errorf("insecure config file permissions %04o (should be %04o or %04o): %s", fileMode, _configFilePerm, _configFilePermReadOnly, filename)
		}
	}

	return nil
}
