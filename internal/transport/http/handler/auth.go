package handler

import (
	"net/http"
	"strconv"

	"my_mdb/internal/service"
	"my_mdb/internal/transport/http/respond"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	app *service.App
}

func NewAuth(app *service.App) *AuthHandler { return &AuthHandler{app: app} }

func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	id, err := h.app.Auth.CreateUser(r.Context())
	if err != nil {
		respond.FromServiceError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusCreated, map[string]int{"user_id": id})
}

func (h *AuthHandler) ValidateUser(w http.ResponseWriter, r *http.Request) {
	raw := chi.URLParam(r, "userID")
	if raw == "" {
		respond.Error(w, r, respond.ErrBadRequest("missing userID"))
		return
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		respond.Error(w, r, respond.ErrBadRequest("invalid userID"))
		return
	}

	if err := h.app.Auth.ValidateUserID(r.Context(), id); err != nil {
		respond.FromServiceError(w, r, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"valid":   true,
		"user_id": id,
	})
}
