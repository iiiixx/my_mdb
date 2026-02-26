package converter

import (
	"my_mdb/internal/domain"
	"my_mdb/internal/service"
	"my_mdb/internal/transport/http/handler/dto"
)

func ToHomePage(in *service.HomePage) dto.HomePage {
	out := dto.HomePage{
		Genres: in.Genres,
		Changing: dto.ChangingBlock{
			Kind:  in.Changing.Kind,
			Title: in.Changing.Title,
		},
	}

	for _, c := range in.ForYou {
		out.ForYou = append(out.ForYou, toCard(c))
	}
	for _, c := range in.Top200Pick {
		out.Top200Pick = append(out.Top200Pick, toCard(c))
	}
	for _, c := range in.Changing.Movies {
		out.Changing.Movies = append(out.Changing.Movies, toCard(c))
	}

	return out
}

func toCard(c domain.MovieCard) dto.MovieCard {
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
