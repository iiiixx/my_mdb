package service

import (
	"my_mdb/internal/config"

	"github.com/sirupsen/logrus"
)

type App struct {
	Auth   *AuthService
	Movies *MoviesService
	Home   *HomeService
	Log    *logrus.Logger
}

type Deps struct {
	Users UsersRepo

	Movies  MoviesRepo
	Ratings RatingsRepo
	Posters PostersRepo
	Details MovieDetailsRepo
	Recs    RecommendationsRepo
	Similar SimilarityRepo
	Tags    TagsRepo

	OMDb OMDbClient
	Rec  RecClient

	Cfg config.Config
	Log *logrus.Logger
}

func NewApp(d Deps) *App {
	moviesvc := NewMoviesService(MoviesServiceDeps{
		Movies:    d.Movies,
		Ratings:   d.Ratings,
		Posters:   d.Posters,
		Details:   d.Details,
		Similar:   d.Similar,
		RecsRepo:  d.Recs,
		RecClient: d.Rec,
		Tags:      d.Tags,
		OMDb:      d.OMDb,
		Cfg:       d.Cfg,
		Log:       d.Log,
	})
	return &App{
		Auth:   NewAuthService(d.Users, d.Log),
		Movies: moviesvc,
		Home: NewHomeService(HomeServiceDeps{
			Movies:  d.Movies,
			Ratings: d.Ratings,
			Tags:    d.Tags,
			Posters: moviesvc,
			Recs:    moviesvc,
			Log:     d.Log,
			Cfg:     d.Cfg,
		}),
		Log: d.Log,
	}
}
