package handler

import (
	"net/http"

	"my_mdb/internal/service"
	"my_mdb/internal/transport/http/handler/converter"
	middleware "my_mdb/internal/transport/http/mymiddleware"
	"my_mdb/internal/transport/http/respond"
)

type HomeHandler struct {
	app *service.App
}

func NewHome(app *service.App) *HomeHandler { return &HomeHandler{app: app} }

func (h *HomeHandler) Home(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromCtx(r.Context())
	if !ok {
		respond.Error(w, r, respond.ErrBadRequest("missing userID in context"))
		return
	}

	page, err := h.app.Home.BuildHome(r.Context(), userID)
	if err != nil {
		respond.FromServiceError(w, r, err)
		return
	}
	out := converter.ToHomePage(page)

	respond.JSON(w, http.StatusOK, out)
}
