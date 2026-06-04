package data

import (
	"atwell.dev/reelingit/models"
)

type MovieStorage interface {
	GetTopMovies() ([]models.Movie, error)
	GetRandomMovies() ([]models.Movie, error)
	GetMovieByID(id int) (models.Movie, error)
	SearchMoviesByName(name string, order string, genre *int) ([]models.Movie, error)
	GetAllGenres() ([]models.Genre, error)
}