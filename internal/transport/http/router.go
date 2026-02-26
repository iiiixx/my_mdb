package http

import (
	"net/http"

	"my_mdb/internal/service"
	"my_mdb/internal/transport/http/handler"
	mymw "my_mdb/internal/transport/http/mymiddleware"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(app *service.App) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(mymw.Logger(app.Log))

	ah := handler.NewAuth(app)
	hh := handler.NewHome(app)
	mh := handler.NewMovies(app)
	rh := handler.NewRatings(app)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Route("/api", func(r chi.Router) {
		r.Post("/users", ah.CreateUser)
		r.Get("/users/{userID}/validate", ah.ValidateUser)

		r.Get("/movies/search", mh.Search)
		r.Get("/movies/top200", mh.Top200)
		r.Get("/movies/genres", mh.Genres)
		r.Get("/movies/genre/{genre}", mh.ByGenre)
		r.Get("/movies/tag", mh.ByTagQuery)

		r.Route("/users/{userID}", func(r chi.Router) {
			r.Use(mymw.UserCtx(app.Auth))

			r.Get("/home", hh.Home)

			r.Get("/watched", mh.WatchedByUser)
			r.Get("/movies/{movieID}", mh.Details)

			r.Put("/ratings/{movieID}", rh.Upsert)
		})
	})

	return r
}
