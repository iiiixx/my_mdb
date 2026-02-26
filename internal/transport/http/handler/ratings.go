package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"my_mdb/internal/service"
	"my_mdb/internal/transport/http/middleware"
	"my_mdb/internal/transport/http/respond"

	"github.com/go-chi/chi/v5"
)

type RatingsHandler struct {
	app *service.App
}

func NewRatings(app *service.App) *RatingsHandler { return &RatingsHandler{app: app} }

type upsertRatingReq struct {
	Value float32 `json:"value"`
}

func (h *RatingsHandler) Upsert(w http.ResponseWriter, r *http.Request) {
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

	var req upsertRatingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, r, respond.ErrBadRequest("invalid json body"))
		return
	}

	if err := h.app.Movies.Rate(r.Context(), userID, movieID, req.Value); err != nil {
		respond.FromServiceError(w, r, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}
