package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"atwell.dev/reelingit/data"
	"atwell.dev/reelingit/handlers"
	"atwell.dev/reelingit/logger"
	"github.com/joho/godotenv"
)

// database connection string:  postgres://postgres:postgres@localhost:5432/reelingit?sslmode=disable

func main() {
	// .env loading
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}

	// initialize logger
	logInstance := initializeLogger()

	// Database connection
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		log.Fatalf("DATABASE_URL not set in environment")
	}
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// initialize movie repo
	movieRepo, err := data.NewMovieRepository(db, logInstance)
	if err != nil {
		log.Fatalf("Failed to initialize movie repository: %v", err)
	}

	// serve static files
	http.Handle("/", http.FileServer(http.Dir("public")))

	// Initialize handlers
	movieHandler := handlers.NewMovieHandler(movieRepo, logInstance)
	// authHandler := handlers.NewAuthHandler(userStorage, jwt, logInstance)

	// Set up routes
	http.HandleFunc("/api/movies/random", movieHandler.GetRandomMovies)
	http.HandleFunc("/api/movies/top", movieHandler.GetTopMovies)
	http.HandleFunc("/api/movies/search", movieHandler.SearchMovies)
	http.HandleFunc("/api/movies/", movieHandler.GetMovie)
	http.HandleFunc("/api/genres", movieHandler.GetGenres)
	http.HandleFunc("/api/account/register", movieHandler.GetGenres)
	http.HandleFunc("/api/account/authenticate", movieHandler.GetGenres)

	// start server
	const addr = ":8080"
	logInstance.Info("Server starting on " + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logInstance.Error("Server failed to start", err)
		log.Fatalf("Server failed: %v", err)
	}
}

func initializeLogger() *logger.Logger {
	logInstance, err := logger.NewLogger("movie-service.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	return logInstance
}