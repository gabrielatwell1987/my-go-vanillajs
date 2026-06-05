package models

type User struct {
	ID             int     `json:"id"`
	Email          string  `json:"email"`
	Name           string  `json:"name"`
	Password       string  `json:"password,omitempty"`
	PasswordHashed string  `json:"-"`
	Favorites      []Movie `json:"favorites,omitempty"`
	Watchlist      []Movie `json:"watchlist,omitempty"`
}