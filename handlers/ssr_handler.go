package handlers

import (
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"atwell.dev/reelingit/data"
	"atwell.dev/reelingit/logger"
	"atwell.dev/reelingit/models"
)

// In main.go, add this new handler function before the main function
func SSRMovieDetailsHandler(movieRepo *data.MovieRepository, logInstance *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract movie ID from URL
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.Error(w, "Movie ID required", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "Invalid movie ID", http.StatusBadRequest)
			return
		}

		// Get movie from repository
		movie, err := movieRepo.GetMovieByID(id)
		if err != nil {
			if errors.Is(err, data.ErrMovieNotFound) {
				http.Error(w, "Movie not found", http.StatusNotFound)
			} else {
				logInstance.Error("Error fetching movie", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Serve the HTML with movie data
		w.Header().Set("Content-Type", "text/html")
		err = renderMovieDetails(w, movie)
		if err != nil {
			logInstance.Error("Error rendering movie details", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// Add this function to render the HTML
func renderMovieDetails(w io.Writer, movie models.Movie) error {
	// Read the index.html file
	htmlContent, err := os.ReadFile("./public/index.html")
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Convert movie data to HTML
	genresHTML := ""
	for _, genre := range movie.Genres {
		genresHTML += fmt.Sprintf(`<li>%s</li>`, html.EscapeString(genre.Name))
	}

	castHTML := ""
	for _, actor := range movie.Casting {
		imageURL := "/images/generic_actor.jpg"
		if actor.ImageURL != nil {
			imageURL = *actor.ImageURL
		}
		castHTML += fmt.Sprintf(`
            <li>
                <img src="%s" alt="Picture of %s">
                <p>%s %s</p>
            </li>`,
			html.EscapeString(imageURL),
			html.EscapeString(actor.LastName),
			html.EscapeString(actor.FirstName),
			html.EscapeString(actor.LastName))
	}

	// Replace the main content
	mainContent := fmt.Sprintf(`
        <main>
            <article id="movie">
                <h2>%s</h2>
                <h3>%s</h3>
                <header>
                    <img src="%s" alt="Poster">
                    <youtube-embed id="trailer" data-url="%s"></youtube-embed>
                    <section id="actions">
                        <dl id="metadata">
                            <dt>Release Date</dt>
                            <dd>%d</dd>
                            <dt>Score</dt>
                            <dd>%.1f / 10</dd>
                            <dt>Original language</dt>
                            <dd>%s</dd>
                        </dl>
                        <button id="btnFavorites">Add to Favorites</button>
                        <button id="btnWatchlist">Add to Watchlist</button>
                    </section>
                </header>
                <ul id="genres">%s</ul>
                <p id="overview">%s</p>
                <ul id="cast">%s</ul>
            </article>
        </main>`,
		html.EscapeString(movie.Title),
		html.EscapeString(*movie.Tagline),
		html.EscapeString(*movie.PosterURL),
		html.EscapeString(*movie.TrailerURL),
		movie.ReleaseYear,
		*movie.Score,
		html.EscapeString(*movie.Language),
		genresHTML,
		html.EscapeString(*movie.Overview),
		castHTML)

	// Replace the main tag content in the HTML
	htmlStr := string(htmlContent)
	htmlStr = strings.Replace(htmlStr, "<main></main>", mainContent, 1)
	fmt.Println(htmlStr)

	// Write the response
	_, err = w.Write([]byte(htmlStr))
	return err
}