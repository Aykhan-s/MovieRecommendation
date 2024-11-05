package config

import (
	"fmt"
	"strconv"

	"github.com/aykhans/movier/server/pkg/utils"
)

type DownloadConfig struct {
	URL          string
	DownloadName string
	ExtractName  string
}

var DownloadConfigs = []DownloadConfig{
	{
		URL:          "https://datasets.imdbws.com/title.basics.tsv.gz",
		DownloadName: "title.basics.tsv.gz",
		ExtractName:  "title.basics.tsv",
	},
	{
		URL:          "https://datasets.imdbws.com/title.principals.tsv.gz",
		DownloadName: "title.principals.tsv.gz",
		ExtractName:  "title.principals.tsv",
	},
	{
		URL:          "https://datasets.imdbws.com/title.ratings.tsv.gz",
		DownloadName: "title.ratings.tsv.gz",
		ExtractName:  "title.ratings.tsv",
	},
}

var BaseDir = "/"

func GetTemplatePath() string {
	return BaseDir + "/pkg/templates"
}

func GetDownloadPath() string {
	return BaseDir + "/data/raw"
}

func GetExtractPath() string {
	return BaseDir + "/data/extracted"
}

const (
	ServePort = 8080
)

var TitleTypes = []string{"movie", "tvMovie"}
var NconstCategories = []string{"actress", "actor", "director", "writer"}

func NewPostgresURL() (string, error) {
	username := utils.GetEnv("POSTGRES_USER", "")
	if username == "" {
		return "", fmt.Errorf("POSTGRES_USER env variable is not set")
	}
	password := utils.GetEnv("POSTGRES_PASSWORD", "")
	if password == "" {
		return "", fmt.Errorf("POSTGRES_PASSWORD env variable is not set")
	}
	host := utils.GetEnv("POSTGRES_HOST", "")
	if host == "" {
		return "", fmt.Errorf("POSTGRES_HOST env variable is not set")
	}
	port := utils.GetEnv("POSTGRES_PORT", "")
	if port == "" {
		return "", fmt.Errorf("POSTGRES_PORT env variable is not set")
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("POSTGRES_PORT env variable is not a number")
	}
	db := utils.GetEnv("POSTGRES_DB", "")
	if db == "" {
		return "", fmt.Errorf("POSTGRES_DB env variable is not set")
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username, password, host, port, db,
	), nil
}

func NewRecommenderServiceGrpcTarget() (string, error) {
	host := utils.GetEnv("RECOMMENDER_SERVICE_GRPC_HOST", "")
	if host == "" {
		return "", fmt.Errorf("RECOMMENDER_SERVICE_GRPC_HOST env variable is not set")
	}
	port := utils.GetEnv("RECOMMENDER_SERVICE_GRPC_PORT", "")
	if port == "" {
		return "", fmt.Errorf("RECOMMENDER_SERVICE_GRPC_PORT env variable is not set")
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("RECOMMENDER_SERVICE_GRPC_PORT env variable is not a number")
	}
	return fmt.Sprintf("%s:%s", host, port), nil
}

func GetBaseURL() string {
	return utils.GetEnv("BASE_URL", "http://localhost:8080")
}
