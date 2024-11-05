package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aykhans/movier/server/pkg/config"
	"github.com/aykhans/movier/server/pkg/handlers"
	"github.com/aykhans/movier/server/pkg/storage/postgresql"
	"github.com/aykhans/movier/server/pkg/storage/postgresql/repository"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getServeCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Movie Recommendation Serve",
		Run: func(cmd *cobra.Command, args []string) {
			err := runServe()
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Movie Recommendation Serve")
		},
	}
	return serveCmd
}

func runServe() error {
	dbURL, err := config.NewPostgresURL()
	if err != nil {
		return err
	}
	db, err := postgresql.NewDB(dbURL)
	defer db.Close(context.Background())
	if err != nil {
		return err
	}
	imdbRepo := repository.NewIMDbRepository(db)

	grpcRecommenderServiceTarget, err := config.NewRecommenderServiceGrpcTarget()
	if err != nil {
		return err
	}
	conn, err := grpc.NewClient(
		grpcRecommenderServiceTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect to grpc recommender service: %v", err)
	}
	defer conn.Close()

	router := http.NewServeMux()
	imdbHandler := handlers.NewIMDbHandler(*imdbRepo, conn, config.GetBaseURL())

	router.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})
	router.HandleFunc("GET /", imdbHandler.HandlerHome)
	router.HandleFunc("GET /recs", imdbHandler.HandlerGetRecommendations)

	log.Printf("serving on port %d", config.ServePort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.ServePort), handlers.CORSMiddleware(router))
	if err != nil {
		return err
	}
	return nil
}
