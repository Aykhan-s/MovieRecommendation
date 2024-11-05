package cmd

import (
	"log"

	"github.com/aykhans/movier/server/pkg/config"
	"github.com/aykhans/movier/server/pkg/dto"
	"github.com/aykhans/movier/server/pkg/utils"
	"github.com/spf13/cobra"
)

func getDownloadCmd() *cobra.Command {
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Movie Data Downloader",
		Run: func(cmd *cobra.Command, args []string) {
			err := runDownload()
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	return downloadCmd
}

func runDownload() error {
	downloadPath := config.GetDownloadPath()
	extractPath := config.GetExtractPath()
	err := utils.MakeDirIfNotExist(downloadPath)
	if err != nil {
		return err
	}
	err = utils.MakeDirIfNotExist(extractPath)
	if err != nil {
		return err
	}
	download(downloadPath, extractPath)
	return nil
}

func download(
	downloadPath string,
	extractPath string,
) error {
	for _, downloadConfig := range config.DownloadConfigs {
		extracted, err := utils.IsDirExist(extractPath + "/" + downloadConfig.ExtractName)
		if err != nil {
			return err
		}
		if extracted {
			log.Printf("File %s already extracted. Skipping...\n\n", downloadConfig.ExtractName)
			continue
		}

		downloaded, err := utils.IsDirExist(downloadPath + "/" + downloadConfig.DownloadName)
		if err != nil {
			return err
		}
		if downloaded {
			log.Printf("File %s already downloaded. Extracting...\n\n", downloadConfig.DownloadName)
			if err := dto.ExtractGzFile(
				downloadPath+"/"+downloadConfig.DownloadName,
				extractPath+"/"+downloadConfig.ExtractName,
			); err != nil {
				return err
			}
			continue
		}

		log.Printf("Downloading and extracting %s file...\n\n", downloadConfig.DownloadName)
		if err := dto.DownloadAndExtractGz(
			downloadConfig.URL,
			downloadPath+"/"+downloadConfig.DownloadName,
			extractPath+"/"+downloadConfig.ExtractName,
		); err != nil {
			return err
		}
	}

	return nil
}
