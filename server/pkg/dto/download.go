package dto

import (
	"io"
	"net/http"
	"os"
)

func DownloadAndExtractGz(url, downloadFilepath, extractFilepath string) error {
	if err := Download(url, downloadFilepath); err != nil {
		return err
	}
	return ExtractGzFile(downloadFilepath, extractFilepath)
}

func Download(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
