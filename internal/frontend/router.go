package frontend

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Frontend struct {
	H          *Handlers
	APIBaseURL string
	Client     *http.Client
}

func New(apiBaseURL string) (*Frontend, error) {
	rend, err := NewRenderer()
	if err != nil {
		return nil, err
	}
	h := &Handlers{R: rend}
	return &Frontend{
		H:          h,
		APIBaseURL: apiBaseURL,
		Client:     &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func (f *Frontend) Routes() chi.Router {
	r := chi.NewRouter()

	// static
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// auth pages
	r.Get("/login", f.H.LoginPage)
	r.Post("/login", f.loginPostReal)
	r.Get("/logout", f.H.Logout)

	// protected
	r.Group(func(pr chi.Router) {
		pr.Use(f.H.RequireAuth)
		pr.Get("/", f.H.Home)
		pr.Get("/top200", f.H.Top200)
		pr.Get("/genres", f.H.Genres)
		pr.Get("/genre/{genre}", f.H.Genre)
		pr.Get("/watched", f.H.Watched)
		pr.Get("/recommended", f.H.Recommended)
		pr.Get("/movie/{movieID}", f.H.Movie)
	})

	return r
}

func (f *Frontend) loginPostReal(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	uidStr := r.FormValue("user_id")
	uid, _ := strconv.Atoi(uidStr)
	if uid <= 0 {
		f.H.R.Render(w, "login.html", PageData{Title: "Login", Error: "invalid user id"})
		return
	}

	url := fmt.Sprintf("%s/api/users/%d/validate", f.APIBaseURL, uid)
	resp, err := f.Client.Get(url)
	if err != nil || resp.StatusCode >= 400 {
		f.H.R.Render(w, "login.html", PageData{Title: "Login", Error: "user not valid"})
		if resp != nil {
			_ = resp.Body.Close()
		}
		return
	}
	_ = resp.Body.Close()

	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: strconv.Itoa(uid), Path: "/", HttpOnly: true})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (f *Frontend) apiPost(path string, body []byte) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodPost, f.APIBaseURL+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return f.Client.Do(req)
}

func (f *Frontend) apiGet(path string) ([]byte, int, error) {
	resp, err := f.Client.Get(f.APIBaseURL + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return b, resp.StatusCode, nil
}
