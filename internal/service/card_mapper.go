package service

import "my_mdb/internal/domain"

func toCard(m domain.Movie, posterURL *string, userRate *float32, recScore *float32) domain.MovieCard {
	return domain.MovieCard{
		ID:        m.ID,
		Title:     m.Title,
		Year:      m.Year,
		Genres:    m.Genres,
		PosterURL: posterURL,
		UserRate:  userRate,
		RecScore:  recScore,
	}
}

func toCards(movies []domain.Movie) []domain.MovieCard {
	out := make([]domain.MovieCard, 0, len(movies))
	for i := range movies {
		out = append(out, toCard(movies[i], nil, nil, nil))
	}
	return out
}

func posterPtr(pm map[int]string, movieID int) *string {
	if pm == nil {
		return nil
	}
	if url, ok := pm[movieID]; ok && url != "" && url != "N/A" {
		u := url
		return &u
	}
	return nil
}

type CardMapOpt struct {
	UserRates map[int]*float32
	RecScores map[int]float32
}

func mapMoviesToCards(mvs []domain.Movie, posterMap map[int]string, opt CardMapOpt) []domain.MovieCard {
	out := make([]domain.MovieCard, 0, len(mvs))
	for i := range mvs {
		mv := mvs[i]

		var rate *float32
		if opt.UserRates != nil {
			rate = opt.UserRates[mv.ID]
		}

		var recScore *float32
		if opt.RecScores != nil {
			if sc, ok := opt.RecScores[mv.ID]; ok {
				s := sc
				recScore = &s
			}
		}

		out = append(out, toCard(mv, posterPtr(posterMap, mv.ID), rate, recScore))
	}
	return out
}
