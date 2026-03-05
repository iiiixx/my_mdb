package frontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	R *Renderer
}

type PageData struct {
	Title   string
	UserID  int
	Error   string
	Heading string
	MovieID int
	Page    string
	Param   string
}

const cookieName = "mdb_uid"

func getUID(r *http.Request) (int, bool) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return 0, false
	}
	id, err := strconv.Atoi(c.Value)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func (h *Handlers) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getUID(r); !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.R.Render(w, "login.html", PageData{Title: "Login", UserID: 0})
}

func (h *Handlers) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.R.Render(w, "login.html", PageData{Title: "Login", Error: "bad form"})
		return
	}
	idStr := r.FormValue("user_id")
	uid, _ := strconv.Atoi(idStr)
	if uid <= 0 {
		h.R.Render(w, "login.html", PageData{Title: "Login", Error: "invalid user id"})
		return
	}

	// дергаем твою ручку validate (в этом же хосте)
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/validate", uid), nil)
	rr := &responseRecorder{header: make(http.Header)}
	http.DefaultServeMux.ServeHTTP(rr, req) // если у тебя не DefaultServeMux — см. ниже "router.go"

	// УПРОЩЕНИЕ: лучше использовать http.Client на http://localhost:8080
	// но мы сделаем правильно в router.go через "apiBaseURL".

	// Поэтому тут просто редиректим (реальную проверку делаем ниже в router.go вариантом через client)
	_ = rr
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: strconv.Itoa(uid), Path: "/", HttpOnly: true})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	uid, _ := getUID(r)
	h.R.Render(w, "home.html", PageData{Title: "Home", UserID: uid})
}

func (h *Handlers) Top200(w http.ResponseWriter, r *http.Request) {
	uid, _ := getUID(r)
	h.R.Render(w, "list.html", PageData{
		Title: "Top 200", Heading: "Топ 200", UserID: uid,
		Page: "top200",
	})
}

func (h *Handlers) Genres(w http.ResponseWriter, r *http.Request) {
	uid, _ := getUID(r)
	h.R.Render(w, "list.html", PageData{
		Title: "Genres", Heading: "Жанры", UserID: uid,
		Page: "genres",
	})
}

func (h *Handlers) Genre(w http.ResponseWriter, r *http.Request) {
	uid, _ := getUID(r)
	genre := chi.URLParam(r, "genre")
	h.R.Render(w, "list.html", PageData{
		Title: genre, Heading: genre, UserID: uid,
		Page:  "genre",
		Param: genre,
	})
}

func (h *Handlers) Watched(w http.ResponseWriter, r *http.Request) {
	uid, _ := getUID(r)
	h.R.Render(w, "list.html", PageData{
		Title: "Watched", Heading: "Просмотренное", UserID: uid,
		Page: "watched",
	})
}

func (h *Handlers) Recommended(w http.ResponseWriter, r *http.Request) {
	uid, _ := getUID(r)
	h.R.Render(w, "list.html", PageData{
		Title: "For You", Heading: "Для вас", UserID: uid,
		Page: "recommended",
	})
}

func (h *Handlers) Movie(w http.ResponseWriter, r *http.Request) {
	uid, _ := getUID(r)
	midStr := chi.URLParam(r, "movieID")
	mid, _ := strconv.Atoi(midStr)
	h.R.Render(w, "movie.html", PageData{Title: "Movie", UserID: uid, MovieID: mid})
}

type responseRecorder struct {
	code   int
	header http.Header
	body   []byte
}

func (r *responseRecorder) Header() http.Header { return r.header }
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}
func (r *responseRecorder) WriteHeader(statusCode int) { r.code = statusCode }

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
