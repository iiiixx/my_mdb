package dto

import "my_mdb/internal/domain"

type HomePage struct {
	ForYou     []MovieCard   `json:"for_you"`
	Top200Pick []MovieCard   `json:"top_200_pick"`
	Genres     []string      `json:"genres"`
	Changing   ChangingBlock `json:"changing"`
}

type ChangingBlock struct {
	Kind   string      `json:"kind"`
	Title  string      `json:"title"`
	Movies []MovieCard `json:"movies"`
}

func FromDomainCard(c domain.MovieCard) MovieCard {
	return MovieCard{
		ID:        c.ID,
		Title:     c.Title,
		Year:      c.Year,
		Genres:    c.Genres,
		PosterURL: c.PosterURL,
		UserRate:  c.UserRate,
		RecScore:  c.RecScore,
	}
}
