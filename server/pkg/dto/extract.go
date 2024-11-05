package dto

import (
	"compress/gzip"
	"io"
	"os"
)

func ExtractGzFile(gzFile, extractedFilepath string) error {
	file, err := os.Open(gzFile)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	outFile, err := os.Create(extractedFilepath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, gzReader)
	return err
}
