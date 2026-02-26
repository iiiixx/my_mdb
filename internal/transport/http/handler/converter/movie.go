package converter

import (
	"my_mdb/internal/domain"
	"my_mdb/internal/service"
	"my_mdb/internal/transport/http/handler/dto"
)

func ToMovieDetailsResponse(in *service.MovieDetailsResponse) dto.MovieDetailsResponse {
	if in == nil {
		return dto.MovieDetailsResponse{}
	}

	out := dto.MovieDetailsResponse{
		Movie: dto.Movie{
			ID:     in.Movie.ID,
			Title:  in.Movie.Title,
			Year:   in.Movie.Year,
			Genres: in.Movie.Genres,
			IMDbID: in.Movie.IMDbID,
		},
		Poster:   in.Poster,
		Details:  in.Details,
		UserRate: in.UserRate,
		Similar:  make([]dto.MovieCard, 0, len(in.Similar)),
	}

	for _, c := range in.Similar {
		out.Similar = append(out.Similar, toMovieCard(c))
	}

	return out
}

func ToMovieCard(c domain.MovieCard) dto.MovieCard {
	return dto.MovieCard{
		ID:        c.ID,
		Title:     c.Title,
		Year:      c.Year,
		Genres:    c.Genres,
		PosterURL: c.PosterURL,
		UserRate:  c.UserRate,
		RecScore:  c.RecScore,
	}
}

func ToMovieCards(in []domain.MovieCard) []dto.MovieCard {
	out := make([]dto.MovieCard, 0, len(in))
	for _, c := range in {
		out = append(out, ToMovieCard(c))
	}
	return out
}
