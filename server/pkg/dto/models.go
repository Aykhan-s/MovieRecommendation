package dto

type Basic struct {
	Tconst    string `json:"tconst"`
	StartYear uint16 `json:"startYear"`
	Genres    string `json:"genres"`
}

type Principal struct {
	Tconst  string `json:"tconst"`
	Nconsts string `json:"nconsts"`
}

type Ratings struct {
	Tconst string  `json:"tconst"`
	Rating float64 `json:"rating"`
	Votes  int     `json:"votes"`
}

type MinMax struct {
	MinVotes  uint    `json:"minVotes"`
	MaxVotes  uint    `json:"maxVotes"`
	MinYear   uint    `json:"minYear"`
	MaxYear   uint    `json:"maxYear"`
	MinRating float64 `json:"minRating"`
	MaxRating float64 `json:"maxRating"`
}
