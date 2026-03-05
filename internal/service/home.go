package service

import (
	"context"

	"my_mdb/internal/config"
	"my_mdb/internal/domain"

	"github.com/sirupsen/logrus"
)

type HomeServiceDeps struct {
	Movies  MoviesRepo
	Ratings RatingsRepo
	Tags    TagsRepo

	Posters PosterProvider
	Recs    RecProvider

	Cfg config.Config
	Log *logrus.Logger
}

type HomeService struct {
	movies  MoviesRepo
	ratings RatingsRepo
	tags    TagsRepo
	posters PosterProvider
	recs    RecProvider
	cfg     config.Config
	log     *logrus.Logger
}

func NewHomeService(d HomeServiceDeps) *HomeService {
	return &HomeService{
		movies:  d.Movies,
		ratings: d.Ratings,
		tags:    d.Tags,
		posters: d.Posters,
		recs:    d.Recs,
		cfg:     d.Cfg,
		log:     d.Log,
	}
}

type HomePage struct {
	ForYou     []domain.MovieCard
	Top200Pick []domain.MovieCard
	Genres     []string
	Changing   domain.ChangingBlockResponse
}

func (s *HomeService) BuildHome(ctx context.Context, userID int) (*HomePage, error) {
	forYouMovies, err := s.recs.GetForYouMovies(ctx, userID, 6)
	if err != nil {
		return nil, err
	}

	var forYou []domain.MovieCard
	if len(forYouMovies) == 0 {
		forYou = []domain.MovieCard{}
	} else {
		var pm map[int]string
		if s.posters != nil {
			pm, err = s.posters.PosterMapForMovies(ctx, forYouMovies, true)
			if err != nil {
				return nil, err
			}
		}
		forYou = mapMoviesToCards(forYouMovies, pm, CardMapOpt{})
	}

	topPickMovies, err := s.movies.RandomFromTop(ctx, 200, 6)
	if err != nil {
		return nil, err
	}
	var topPickCards []domain.MovieCard
	if s.posters != nil {
		pm, err := s.posters.PosterMapForMovies(ctx, topPickMovies, true)
		if err != nil {
			return nil, err
		}
		topPickCards = mapMoviesToCards(topPickMovies, pm, CardMapOpt{})
	}

	genres, err := s.movies.ListGenres(ctx)
	if err != nil {
		return nil, err
	}

	pick, err := s.pickChangingBlockDaily(ctx, userID, 6)
	if err != nil {
		return nil, err
	}

	pm, err := s.posters.PosterMapForMovies(ctx, pick.Movies, true)
	if err != nil {
		return nil, err
	}

	changingCards := mapMoviesToCards(pick.Movies, pm, CardMapOpt{})

	return &HomePage{
		ForYou:     forYou,
		Top200Pick: topPickCards,
		Genres:     genres,
		Changing: domain.ChangingBlockResponse{
			Kind:   pick.Kind,
			Title:  pick.Title,
			Movies: changingCards,
		},
	}, nil
}
