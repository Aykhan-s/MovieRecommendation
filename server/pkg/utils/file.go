package utils

import "os"

func MakeDirIfNotExist(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func IsDirExist(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}
