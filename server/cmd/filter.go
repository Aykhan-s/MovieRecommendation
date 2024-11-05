package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/aykhans/movier/server/pkg/config"
	"github.com/aykhans/movier/server/pkg/dto"

	"github.com/aykhans/movier/server/pkg/storage/postgresql"
	"github.com/aykhans/movier/server/pkg/storage/postgresql/repository"
	"github.com/spf13/cobra"
)

func getFilterCmd() *cobra.Command {
	filterCmd := &cobra.Command{
		Use:   "filter",
		Short: "Movie Data Filter",
		Run: func(cmd *cobra.Command, args []string) {
			err := runFilter()
			if err != nil {
				log.Fatalln(err)
			}
		},
	}

	return filterCmd
}

func runFilter() error {
	generalStartTime := time.Now()
	extractedPath := config.GetExtractPath()

	log.Printf("Filtering basics data...\n\n")
	startTime := time.Now()
	basics, err := dto.FilterBasics(extractedPath + "/title.basics.tsv")
	if err != nil {
		return err
	}
	log.Printf("Basics data filtered. Found %d records (%s)\n\n", len(basics), time.Since(startTime))

	log.Printf("Inserting basics data...\n\n")
	postgresURL, err := config.NewPostgresURL()
	if err != nil {
		return err
	}

	db, err := postgresql.NewDB(postgresURL)
	if err != nil {
		return err
	}
	imdbRepo := repository.NewIMDbRepository(db)
	startTime = time.Now()
	err = imdbRepo.InsertMultipleBasics(basics)
	if err != nil {
		return err
	}
	log.Printf("Basics data inserted. (%s)\n\n", time.Since(startTime))

	log.Printf("Filtering principals data...\n\n")
	tconsts, err := imdbRepo.GetAllTconsts()
	if err != nil {
		return err
	}
	if len(tconsts) == 0 {
		return fmt.Errorf("no tconsts found")
	}
	startTime = time.Now()
	principals, err := dto.FilterPrincipals(extractedPath+"/title.principals.tsv", tconsts)
	if err != nil {
		return err
	}
	log.Printf("Principals data filtered. (%s)\n\n", time.Since(startTime))

	log.Printf("Inserting principals data...\n\n")
	startTime = time.Now()
	err = imdbRepo.UpdateMultiplePrincipals(principals)
	if err != nil {
		return err
	}
	log.Printf("Principals data inserted. (%s)\n\n", time.Since(startTime))

	log.Printf("Filtering ratings data...\n\n")
	startTime = time.Now()
	ratings, err := dto.FilterRatings(extractedPath+"/title.ratings.tsv", tconsts)
	if err != nil {
		return err
	}
	log.Printf("Ratings data filtered. (%s)\n\n", time.Since(startTime))

	log.Printf("Inserting ratings data...\n\n")
	startTime = time.Now()
	err = imdbRepo.UpdateMultipleRatings(ratings)
	if err != nil {
		return err
	}
	log.Printf("Ratings data inserted. (%s)\n\n", time.Since(startTime))

	log.Printf("Filtering done! (%s)\n", time.Since(generalStartTime))
	return nil
}
