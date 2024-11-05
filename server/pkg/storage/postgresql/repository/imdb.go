package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/aykhans/movier/server/pkg/dto"
)

type IMDbRepository struct {
	db *pgx.Conn
}

func NewIMDbRepository(db *pgx.Conn) *IMDbRepository {
	return &IMDbRepository{
		db: db,
	}
}

func (repo *IMDbRepository) InsertMultipleBasics(basics []dto.Basic) error {
	batch := &pgx.Batch{}
	for _, basic := range basics {
		batch.Queue(
			`INSERT INTO imdb (tconst, year, genres) 
			VALUES ($1, $2, $3)
			ON CONFLICT (tconst) DO UPDATE 
			SET year = EXCLUDED.year, genres = EXCLUDED.genres`,
			basic.Tconst, basic.StartYear, basic.Genres,
		)
	}

	results := repo.db.SendBatch(context.Background(), batch)
	if err := results.Close(); err != nil {
		return err
	}
	return nil
}

func (repo *IMDbRepository) GetAllTconsts() ([]string, error) {
	rows, err := repo.db.Query(
		context.Background(),
		"SELECT tconst FROM imdb",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tconsts []string
	for rows.Next() {
		var tconst string
		if err := rows.Scan(&tconst); err != nil {
			return nil, err
		}
		tconsts = append(tconsts, tconst)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tconsts, nil
}

func (repo *IMDbRepository) UpdateMultiplePrincipals(principals []dto.Principal) error {
	batch := &pgx.Batch{}
	for _, principal := range principals {
		batch.Queue(
			`UPDATE imdb SET nconsts = $1 WHERE tconst = $2`,
			principal.Nconsts, principal.Tconst,
		)
	}

	results := repo.db.SendBatch(context.Background(), batch)
	if err := results.Close(); err != nil {
		return err
	}
	return nil
}

func (repo *IMDbRepository) UpdateMultipleRatings(ratings []dto.Ratings) error {
	batch := &pgx.Batch{}
	for _, rating := range ratings {
		batch.Queue(
			`UPDATE imdb SET rating = $1, votes = $2 WHERE tconst = $3`,
			rating.Rating, rating.Votes, rating.Tconst,
		)
	}

	results := repo.db.SendBatch(context.Background(), batch)
	if err := results.Close(); err != nil {
		return err
	}
	return nil
}

func (repo *IMDbRepository) GetMinMax() (*dto.MinMax, error) {
	var minMax dto.MinMax

	err := repo.db.QueryRow(
		context.Background(),
		"SELECT MIN(votes), MAX(votes), MIN(year), MAX(year), MIN(rating), MAX(rating) FROM imdb LIMIT 1",
	).Scan(&minMax.MinVotes, &minMax.MaxVotes, &minMax.MinYear, &minMax.MaxYear, &minMax.MinRating, &minMax.MaxRating)
	if err != nil {
		return nil, err
	}
	return &minMax, nil
}
