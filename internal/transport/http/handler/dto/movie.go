package dto

import "encoding/json"

type MovieDetailsResponse struct {
	Movie        Movie           `json:"movie"`
	Poster       *string         `json:"poster_url,omitempty"`
	Details      json.RawMessage `json:"details,omitempty"`
	UserRate     *float32        `json:"user_rating,omitempty"`
	PlatformRate *float32        `json:"platform_rating,omitempty"`
	Similar      []MovieCard     `json:"similar"`
}

type Movie struct {
	ID     int      `json:"id"`
	Title  string   `json:"title"`
	Year   *int16   `json:"year"`
	Genres []string `json:"genres"`
	IMDbID *int     `json:"imdb_id,omitempty"`
}

type MovieCard struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Year      *int16   `json:"year"`
	Genres    []string `json:"genres"`
	PosterURL *string  `json:"poster_url,omitempty"`
	UserRate  *float32 `json:"user_rating,omitempty"`
	RecScore  *float32 `json:"rec_score,omitempty"`
}
