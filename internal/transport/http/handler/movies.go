package handler

import (
	"net/http"
	"strconv"

	"my_mdb/internal/service"
	"my_mdb/internal/transport/http/handler/converter"
	middleware "my_mdb/internal/transport/http/mymiddleware"
	"my_mdb/internal/transport/http/respond"

	"github.com/go-chi/chi/v5"
)

type MoviesHandler struct {
	app *service.App
}

func NewMovies(app *service.App) *MoviesHandler { return &MoviesHandler{app: app} }

func (h *MoviesHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := parseInt(r.URL.Query().Get("limit"), 20)

	res, err := h.app.Movies.Search(r.Context(), q, limit)
	if err != nil {
		respond.FromServiceError(w, r, err)
		return
	}
	out := converter.ToMovieCards(res)

	respond.JSON(w, http.StatusOK, out)
}

func (h *MoviesHandler) Top200(w http.ResponseWriter, r *http.Request) {
	res, err := h.app.Movies.Top200(r.Context())
	if err != nil {
		respond.FromServiceError(w, r, err)
		return
	}
	out := converter.ToMovieCards(res)

	respond.JSON(w, http.StatusOK, out)
}

func (h *MoviesHandler) Genres(w http.ResponseWriter, r *http.Request) {
	res, err := h.app.Movies.Genres(r.Context())
	if err != nil {
		respond.FromServiceError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, res)
}

func (h *MoviesHandler) ByGenre(w http.ResponseWriter, r *http.Request) {
	genre := chi.URLParam(r, "genre")
	limit := parseInt(r.URL.Query().Get("limit"), 20)

	res, err := h.app.Movies.ByGenre(r.Context(), genre, limit)
	if err != nil {
		respond.FromServiceError(w, r, err)
		return
	}
	out := converter.ToMovieCards(res)

	respond.JSON(w, http.StatusOK, out)
}

func (h *MoviesHandler) ByTagQuery(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := parseInt(r.URL.Query().Get("limit"), 20)

	res, err := h.app.Movies.ByTagQuery(r.Context(), q, limit)
	if err != nil {
		respond.FromServiceError(w, r, err)
		return
	}
	out := converter.ToMovieCards(res)

	respond.JSON(w, http.StatusOK, out)
}

func (h *MoviesHandler) WatchedByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromCtx(r.Context())
	if !ok {
		respond.Error(w, r, respond.ErrBadRequest("missing userID in context"))
		return
	}

	limit := parseInt(r.URL.Query().Get("limit"), 50)
	res, err := h.app.Movies.WatchedByUser(r.Context(), userID, limit)
	if err != nil {
		respond.FromServiceError(w, r, err)
		return
	}
	out := converter.ToMovieCards(res)

	respond.JSON(w, http.StatusOK, out)
}

func (h *MoviesHandler) Details(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromCtx(r.Context())
	if !ok {
		respond.Error(w, r, respond.ErrBadRequest("missing userID in context"))
		return
	}

	movieID, err := strconv.Atoi(chi.URLParam(r, "movieID"))
	if err != nil {
		respond.Error(w, r, respond.ErrBadRequest("invalid movieID"))
		return
	}

	res, err := h.app.Movies.GetMovieDetails(r.Context(), userID, movieID)
	if err != nil {
		respond.FromServiceError(w, r, err)
		return
	}
	out := converter.ToMovieDetailsResponse(res)

	respond.JSON(w, http.StatusOK, out)
}

func parseInt(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
