package domain

import "time"

type MovieDetailsCache struct {
	MovieID   int       `json:"movie_id"`
	Payload   []byte    `json:"payload"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Poster struct {
	MovieID   int       `json:"movie_id"`
	PosterURL string    `json:"poster_url"`
	UpdatedAt time.Time `json:"updated_at"`
}
