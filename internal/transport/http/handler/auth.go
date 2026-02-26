package handler

import (
	"net/http"

	"my_mdb/internal/service"
	"my_mdb/internal/transport/http/respond"
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
	// validate уже делает UserCtx middleware, но этот эндпоинт можно оставить
	respond.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}
