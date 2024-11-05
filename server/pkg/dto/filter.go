package dto

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/aykhans/movier/server/pkg/config"
)

func FilterBasics(filePath string) ([]Basic, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	columnCount := 9
	var headers []string
	if scanner.Scan() {
		headers = strings.Split(scanner.Text(), "\t")
		if len(headers) != columnCount {
			return nil, fmt.Errorf("expected %d column headers, found %d", columnCount, len(headers))
		}
	} else {
		return nil, fmt.Errorf("could not read column headers: %v", scanner.Err())
	}

	var (
		tconstIndex    int = -1
		titleTypeIndex int = -1
		startYearIndex int = -1
		genresIndex    int = -1
	)
	for i, header := range headers {
		switch header {
		case "tconst":
			tconstIndex = i
		case "titleType":
			titleTypeIndex = i
		case "startYear":
			startYearIndex = i
		case "genres":
			genresIndex = i
		}
	}
	switch {
	case tconstIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`tconst`")
	case titleTypeIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`titleType`")
	case startYearIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`startYear`")
	case genresIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`genres`")
	}

	var basics []Basic
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, "\t")
		if len(columns) != columnCount {
			fmt.Println("Columns are:", columns)
			return nil, fmt.Errorf("expected %d columns, found %d", columnCount, len(columns))
		}

		if slices.Contains(config.TitleTypes, columns[titleTypeIndex]) {
			var startYearUint16 uint16
			startYear, err := strconv.Atoi(columns[startYearIndex])
			if err != nil {
				startYearUint16 = 0
			} else {
				startYearUint16 = uint16(startYear)
			}

			var genres string
			if columns[genresIndex] == "\\N" {
				genres = ""
			} else {
				genres = strings.ReplaceAll(strings.ToLower(columns[genresIndex]), " ", "")
			}

			basics = append(basics, Basic{
				Tconst:    columns[tconstIndex],
				StartYear: startYearUint16,
				Genres:    genres,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return basics, nil
}

func FilterPrincipals(filePath string, tconsts []string) ([]Principal, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	columnCount := 6
	var headers []string
	if scanner.Scan() {
		headers = strings.Split(scanner.Text(), "\t")
		if len(headers) != columnCount {
			return nil, fmt.Errorf("expected %d column headers, found %d", columnCount, len(headers))
		}
	} else {
		return nil, fmt.Errorf("could not read column headers: %v", scanner.Err())
	}

	var (
		tconstIndex   int = -1
		nconstIndex   int = -1
		categoryIndex int = -1
	)
	for i, header := range headers {
		switch header {
		case "tconst":
			tconstIndex = i
		case "nconst":
			nconstIndex = i
		case "category":
			categoryIndex = i
		}
	}
	switch {
	case tconstIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`tconst`")
	case nconstIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`nconst`")
	case categoryIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`category`")
	}

	tconstMap := make(map[string][]string)
	for _, tconst := range tconsts {
		tconstMap[tconst] = []string{}
	}
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, "\t")
		if len(columns) != columnCount {
			fmt.Println("Columns are:", columns)
			return nil, fmt.Errorf("expected %d columns, found %d", columnCount, len(columns))
		}

		if slices.Contains(config.NconstCategories, columns[categoryIndex]) {
			if _, ok := tconstMap[columns[tconstIndex]]; ok {
				tconstMap[columns[tconstIndex]] = append(tconstMap[columns[tconstIndex]], columns[nconstIndex])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var principals []Principal
	for tconst, nconsts := range tconstMap {
		principals = append(principals, Principal{
			Tconst:  tconst,
			Nconsts: strings.Join(nconsts, ","),
		})
	}
	return principals, nil
}

func FilterRatings(filePath string, tconsts []string) ([]Ratings, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	columnCount := 3
	var headers []string
	if scanner.Scan() {
		headers = strings.Split(scanner.Text(), "\t")
		if len(headers) != columnCount {
			return nil, fmt.Errorf("expected %d column headers, found %d", columnCount, len(headers))
		}
	} else {
		return nil, fmt.Errorf("could not read column headers: %v", scanner.Err())
	}

	var (
		tconstIndex        int = -1
		averageRatingIndex int = -1
		numVotesIndex      int = -1
	)
	for i, header := range headers {
		switch header {
		case "tconst":
			tconstIndex = i
		case "averageRating":
			averageRatingIndex = i
		case "numVotes":
			numVotesIndex = i
		}
	}
	switch {
	case tconstIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`tconst`")
	case averageRatingIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`averageRating`")
	case numVotesIndex == -1:
		return nil, fmt.Errorf("column %s not found", "`numVotes`")
	}

	tconstMap := make(map[string][]string)
	for _, tconst := range tconsts {
		tconstMap[tconst] = []string{}
	}
	var ratings []Ratings
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, "\t")
		if len(columns) != columnCount {
			fmt.Println("Columns are:", columns)
			return nil, fmt.Errorf("expected %d columns, found %d", columnCount, len(columns))
		}

		if _, ok := tconstMap[columns[tconstIndex]]; ok {
			rating, err := strconv.ParseFloat(columns[averageRatingIndex], 32)
			if err != nil {
				rating = 0
			}

			votes, err := strconv.Atoi(columns[numVotesIndex])
			if err != nil {
				votes = 0
			}

			ratings = append(ratings, Ratings{
				Tconst: columns[tconstIndex],
				Rating: math.Round(rating*10) / 10,
				Votes:  votes,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ratings, nil
}
