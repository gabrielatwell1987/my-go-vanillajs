package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"atwell.dev/reelingit/data"
	"atwell.dev/reelingit/handlers"
	"atwell.dev/reelingit/logger"
	"github.com/go-webauthn/webauthn/webauthn"
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

	accountRepo, err := data.NewAccountRepository(db, logInstance)
	if err != nil {
		log.Fatalf("Failed to initialize account repository: %v", err)
	}

	// serve static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // For API routes, don't serve index.html
        if strings.HasPrefix(r.URL.Path, "/api/") {
            http.NotFound(w, r)
            return
        }
        // For everything else, serve the file if it exists, otherwise fall back to index.html
        filePath := "public" + r.URL.Path
        if _, err := os.Stat(filePath); err == nil {
            http.FileServer(http.Dir("public")).ServeHTTP(w, r)
        } else {
            http.ServeFile(w, r, "public/index.html")
        }
    })

	// Initialize handlers
	movieHandler := handlers.NewMovieHandler(movieRepo, logInstance)
	accountHandler := handlers.NewAccountHandler(accountRepo, logInstance)
	// authHandler := handlers.NewAuthHandler(userStorage, jwt, logInstance)

	// Set up routes
	http.HandleFunc("/api/movies/random", movieHandler.GetRandomMovies)
	http.HandleFunc("/api/movies/top", movieHandler.GetTopMovies)
	http.HandleFunc("/api/movies/search", movieHandler.SearchMovies)
	http.HandleFunc("/api/movies/", movieHandler.GetMovie)
	http.HandleFunc("/movies/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Count(r.URL.Path, "/") == 2 && strings.HasPrefix(r.URL.Path, "/movies/") {
			handlers.SSRMovieDetailsHandler(movieRepo, logInstance)(w, r)
		} else {
			catchAllHandler(w, r)
		}
	})
	http.HandleFunc("/api/genres", movieHandler.GetGenres)
	http.HandleFunc("/api/account/register", movieHandler.GetGenres)
	http.HandleFunc("/api/account/authenticate", movieHandler.GetGenres)
	http.HandleFunc("/api/account/register/", accountHandler.Register)
	http.HandleFunc("/api/account/authenticate/", accountHandler.Authenticate)

	http.Handle("/api/account/favorites/",
		accountHandler.AuthMiddleware(http.HandlerFunc(accountHandler.GetFavorites)))
	http.Handle("/api/account/watchlist/",
		accountHandler.AuthMiddleware(http.HandlerFunc(accountHandler.GetWatchlist)))
	http.Handle("/api/account/save-to-collection/",
		accountHandler.AuthMiddleware(http.HandlerFunc(accountHandler.SaveToCollection)))

	// WebAuthn Handlers
	wconfig := &webauthn.Config{
		RPDisplayName: "ReelingIt",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	}

	var webAuthnManager *webauthn.WebAuthn

	if webAuthnManager, err = webauthn.New(wconfig); err != nil {
		logInstance.Error("Error creating WebAuthn", err)
	}

	if err != nil {
		logInstance.Error("Error initialing Passkey engine", err)
	}

	passkeyRepo := data.NewPasskeyRepository(db, *logInstance)
	webAuthnHandler := handlers.NewWebAuthnHandler(passkeyRepo, logInstance, webAuthnManager)
	http.Handle("/api/passkey/registration-begin",
		accountHandler.AuthMiddleware(http.HandlerFunc(webAuthnHandler.WebAuthnRegistrationBeginHandler)))
	http.Handle("/api/passkey/registration-end",
		accountHandler.AuthMiddleware(http.HandlerFunc(webAuthnHandler.WebAuthnRegistrationEndHandler)))
	http.HandleFunc("/api/passkey/authentication-begin", webAuthnHandler.WebAuthnAuthenticationBeginHandler)
	http.HandleFunc("/api/passkey/authentication-end", webAuthnHandler.WebAuthnAuthenticationEndHandler)


	// start server
	const addr = ":8080"
	logInstance.Info("Server starting on " + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logInstance.Error("Server failed to start", err)
		log.Fatalf("Server failed: %v", err)
	}
}

func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}

func initializeLogger() *logger.Logger {
	logInstance, err := logger.NewLogger("movie-service.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	return logInstance
}