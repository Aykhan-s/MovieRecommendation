package utils

import (
	"math"
	"path/filepath"
	"strconv"
)

func IsValidPath(path string) bool {
	return filepath.IsAbs(path)
}

func IsUint32(value int) bool {
	return value >= 0 && value <= math.MaxUint32
}

func IsInt(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}
