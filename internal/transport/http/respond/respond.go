package respond

import (
	"encoding/json"
	"errors"
	"net/http"

	appErr "my_mdb/internal/errors"
)

type apiError struct {
	Error string `json:"error"`
}

type httpErr struct {
	code int
	msg  string
}

func (e httpErr) Error() string { return e.msg }

func ErrBadRequest(msg string) error { return httpErr{code: http.StatusBadRequest, msg: msg} }

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	he := httpErr{code: http.StatusInternalServerError, msg: "internal error"}
	var h httpErr
	if errors.As(err, &h) {
		he = h
	}
	JSON(w, he.code, apiError{Error: he.msg})
}

func FromServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, appErr.ErrBadRequest):
		Error(w, r, httpErr{code: http.StatusBadRequest, msg: err.Error()})
	case errors.Is(err, appErr.ErrNotFound):
		Error(w, r, httpErr{code: http.StatusNotFound, msg: "not found"})
	default:
		Error(w, r, err)
	}
}
