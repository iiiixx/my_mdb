package domain

import "encoding/json"

type Movie struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Year   *int16   `json:"year,omitempty"`
	Genres []string `json:"genres"`
	IMDbID *int     `json:"imdb_id,omitempty"`
	TMDBID *int     `json:"tmdb_id,omitempty"`
}

type MovieCard struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Year      *int16   `json:"year,omitempty"`
	Genres    []string `json:"genres"`
	PosterURL *string  `json:"poster_url,omitempty"`

	RecScore *float32 `json:"rec_score,omitempty"`
	UserRate *float32 `json:"user_rating,omitempty"`
}

type MovieMeta struct {
	Poster  *string
	Details json.RawMessage
}

type MovieDetailsResponse struct {
	Movie        Movie
	Poster       *string
	Details      json.RawMessage
	UserRate     *float32
	PlatformRate *float32
	Similar      []MovieCard
}
