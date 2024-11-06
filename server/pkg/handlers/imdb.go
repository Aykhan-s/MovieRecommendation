package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/aykhans/movier/server/pkg/dto"
	"github.com/aykhans/movier/server/pkg/proto"
	"github.com/aykhans/movier/server/pkg/storage/postgresql/repository"
	"github.com/aykhans/movier/server/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IMDbHandler struct {
	imdbRepo               repository.IMDbRepository
	grpcRecommenderService *grpc.ClientConn
	baseURL                string
}

func NewIMDbHandler(imdbRepo repository.IMDbRepository, grpcRecommenderService *grpc.ClientConn, baseURL string) *IMDbHandler {
	return &IMDbHandler{
		imdbRepo:               imdbRepo,
		grpcRecommenderService: grpcRecommenderService,
		baseURL:                baseURL,
	}
}

func (h *IMDbHandler) HandlerGetRecommendations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	tconstsQ := query["tconst"]
	tconstsLen := len(tconstsQ)
	if tconstsLen < 1 || tconstsLen > 5 {
		RespondWithJSON(w, ErrorResponse{Error: "tconsts should be between 1 and 5"}, http.StatusBadRequest)
		return
	}

	uniqueTconsts := make(map[string]struct{})
	for _, str := range tconstsQ {
		uniqueTconsts[str] = struct{}{}
	}

	invalidTconsts := []string{}
	tconsts := []string{}
	for tconst := range uniqueTconsts {
		tconstLength := len(tconst)
		if 9 > tconstLength || tconstLength > 12 || !strings.HasPrefix(tconst, "tt") {
			invalidTconsts = append(invalidTconsts, tconst)
		}
		tconsts = append(tconsts, tconst)
	}
	if len(invalidTconsts) > 0 {
		RespondWithJSON(
			w,
			ErrorResponse{
				Error: fmt.Sprintf("Invalid tconsts: %s", strings.Join(invalidTconsts, ", ")),
			},
			http.StatusBadRequest,
		)
		return
	}

	n := 5
	nQuery := query.Get("n")
	if nQuery != "" {
		nInt, err := strconv.Atoi(nQuery)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "n should be an integer"}, http.StatusBadRequest)
			return
		}
		if nInt < 1 || nInt > 20 {
			RespondWithJSON(w, ErrorResponse{Error: "n should be greater than 0 and less than 21"}, http.StatusBadRequest)
			return
		}
		n = nInt
	}

	filter := &proto.Filter{}
	minVotesQ := query.Get("min_votes")
	if minVotesQ != "" {
		minVotesInt, err := strconv.Atoi(minVotesQ)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "min_votes should be an integer"}, http.StatusBadRequest)
			return
		}
		if !utils.IsUint32(minVotesInt) {
			RespondWithJSON(w, ErrorResponse{Error: "min_votes should be greater than or equal to 0 and less than or equal to 4294967295"}, http.StatusBadRequest)
			return
		}
		filter.MinVotesOneof = &proto.Filter_MinVotes{MinVotes: uint32(minVotesInt)}
	}

	maxVotesQ := query.Get("max_votes")
	if maxVotesQ != "" {
		maxVotesInt, err := strconv.Atoi(maxVotesQ)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "max_votes should be an integer"}, http.StatusBadRequest)
			return
		}
		if !utils.IsUint32(maxVotesInt) {
			RespondWithJSON(w, ErrorResponse{Error: "max_votes should be greater than 0 or equal to and less than or equal to 4294967295"}, http.StatusBadRequest)
			return
		}
		if uint32(maxVotesInt) < filter.GetMinVotes() {
			RespondWithJSON(w, ErrorResponse{Error: "max_votes should be greater than min_votes"}, http.StatusBadRequest)
			return
		}
		filter.MaxVotesOneof = &proto.Filter_MaxVotes{MaxVotes: uint32(maxVotesInt)}
	}

	minRatingQ := query.Get("min_rating")
	if minRatingQ != "" {
		minRatingFloat, err := strconv.ParseFloat(minRatingQ, 32)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "min_rating should be a float"}, http.StatusBadRequest)
			return
		}
		if minRatingFloat < 0 || minRatingFloat > 10 {
			RespondWithJSON(w, ErrorResponse{Error: "min_rating should be greater than or equal to 0.0 and less than equal to 10.0"}, http.StatusBadRequest)
			return
		}
		filter.MinRatingOneof = &proto.Filter_MinRating{MinRating: float32(minRatingFloat)}
	}

	maxRatingQ := query.Get("max_rating")
	if maxRatingQ != "" {
		maxRatingFloat, err := strconv.ParseFloat(maxRatingQ, 32)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "max_rating should be a float"}, http.StatusBadRequest)
			return
		}
		if maxRatingFloat < 0 || maxRatingFloat > 10 {
			RespondWithJSON(w, ErrorResponse{Error: "max_rating should be greater than or equal to 0.0 and less than or equal to 10.0"}, http.StatusBadRequest)
			return
		}
		if float32(maxRatingFloat) < filter.GetMinRating() {
			RespondWithJSON(w, ErrorResponse{Error: "max_rating should be greater than min_rating"}, http.StatusBadRequest)
			return
		}
		filter.MaxRatingOneof = &proto.Filter_MaxRating{MaxRating: float32(maxRatingFloat)}
	}

	minYearQ := query.Get("min_year")
	if minYearQ != "" {
		minYearInt, err := strconv.Atoi(minYearQ)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "min_year should be an integer"}, http.StatusBadRequest)
			return
		}
		if !utils.IsUint32(minYearInt) {
			RespondWithJSON(w, ErrorResponse{Error: "min_year should be greater than or equal to 0 and less than or equal to 4294967295"}, http.StatusBadRequest)
			return
		}
		filter.MinYearOneof = &proto.Filter_MinYear{MinYear: uint32(minYearInt)}
	}

	maxYearQ := query.Get("max_year")
	if maxYearQ != "" {
		maxYearInt, err := strconv.Atoi(maxYearQ)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "max_year should be an integer"}, http.StatusBadRequest)
			return
		}
		if !utils.IsUint32(maxYearInt) {
			RespondWithJSON(w, ErrorResponse{Error: "max_year should be greater than or equal to 0 and less than or equal to 4294967295"}, http.StatusBadRequest)
			return
		}
		if uint32(maxYearInt) < filter.GetMinYear() {
			RespondWithJSON(w, ErrorResponse{Error: "max_year should be greater than min_year"}, http.StatusBadRequest)
			return
		}
		filter.MaxYearOneof = &proto.Filter_MaxYear{MaxYear: uint32(maxYearInt)}
	}

	yearWeightQ := query.Get("year_weight")
	ratingWeightQ := query.Get("rating_weight")
	genresWeightQ := query.Get("genres_weight")
	nconstsWeightQ := query.Get("nconsts_weight")

	weight := &proto.Weight{}

	features := []string{}
	totalSum := 0
	if yearWeightQ != "" {
		yearWeight, err := strconv.Atoi(yearWeightQ)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "year_weight should be an integer"}, http.StatusBadRequest)
			return
		}
		if yearWeight < 0 || yearWeight > 400 {
			RespondWithJSON(w, ErrorResponse{Error: "year_weight should be greater than or equal to 0 and less than or equal to 400"}, http.StatusBadRequest)
			return
		}
		if yearWeight > 0 {
			weight.Year = uint32(yearWeight)
			totalSum += yearWeight
			features = append(features, "year")
		}
	}
	if ratingWeightQ != "" {
		ratingWeight, err := strconv.Atoi(ratingWeightQ)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "rating_weight should be an integer"}, http.StatusBadRequest)
			return
		}
		if ratingWeight < 0 || ratingWeight > 400 {
			RespondWithJSON(w, ErrorResponse{Error: "rating_weight should be greater than or equal to 0 and less than or equal to 400"}, http.StatusBadRequest)
			return
		}
		if ratingWeight > 0 {
			weight.Rating = uint32(ratingWeight)
			totalSum += ratingWeight
			features = append(features, "rating")
		}
	}
	if genresWeightQ != "" {
		genresWeight, err := strconv.Atoi(genresWeightQ)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "genres_weight should be an integer"}, http.StatusBadRequest)
			return
		}
		if genresWeight < 0 || genresWeight > 400 {
			RespondWithJSON(w, ErrorResponse{Error: "genres_weight should be greater than or equal to 0 and less than or equal to 400"}, http.StatusBadRequest)
			return
		}
		if genresWeight > 0 {
			weight.Genres = uint32(genresWeight)
			totalSum += genresWeight
			features = append(features, "genres")
		}
	}
	if nconstsWeightQ != "" {
		nconstsWeight, err := strconv.Atoi(nconstsWeightQ)
		if err != nil {
			RespondWithJSON(w, ErrorResponse{Error: "nconsts_weight should be an integer"}, http.StatusBadRequest)
			return
		}
		if nconstsWeight < 0 || nconstsWeight > 400 {
			RespondWithJSON(w, ErrorResponse{Error: "nconsts_weight should be greater than or equal to 0 and less than or equal to 400"}, http.StatusBadRequest)
			return
		}
		if nconstsWeight > 0 {
			weight.Nconsts = uint32(nconstsWeight)
			totalSum += nconstsWeight
			features = append(features, "nconsts")
		}
	}

	featuresLen := len(features)
	if featuresLen < 1 {
		RespondWithJSON(w, ErrorResponse{Error: "At least one feature should be selected"}, http.StatusBadRequest)
		return
	}
	if featuresLen*100 != totalSum {
		RespondWithJSON(w, ErrorResponse{Error: fmt.Sprintf("Sum of the %d features should be equal to %d", featuresLen, featuresLen*100)}, http.StatusBadRequest)
		return
	}

	client := proto.NewRecommenderClient(h.grpcRecommenderService)
	response, err := client.GetRecommendations(r.Context(), &proto.Request{
		Tconsts: tconsts,
		N:       uint32(n),
		Filter:  filter,
		Weight:  weight,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				RespondWithJSON(w, ErrorResponse{Error: st.Message()}, http.StatusBadRequest)
			case codes.NotFound:
				RespondWithJSON(w, ErrorResponse{Error: st.Message()}, http.StatusNotFound)
			case codes.Internal:
				RespondWithServerError(w)
			default:
				fmt.Println(err)
				RespondWithServerError(w)
			}
			return
		}
		RespondWithServerError(w)
		return
	}

	RespondWithJSON(w, response.Movies, http.StatusOK)
}

func (h *IMDbHandler) HandlerHome(w http.ResponseWriter, r *http.Request) {
	minMax, err := h.imdbRepo.GetMinMax()
	if err != nil {
		log.Printf("error getting min max: %v", err)
		RespondWithServerError(w)
		return
	}

	RespondWithHTML(
		w, "index.html",
		struct {
			MinMax  dto.MinMax
			BaseURL string
		}{*minMax, h.baseURL},
		http.StatusOK,
	)
}
