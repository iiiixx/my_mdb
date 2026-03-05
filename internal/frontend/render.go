package frontend

import (
	"html/template"
	"net/http"
	"path/filepath"
	"sync"
)

type Renderer struct {
	mu    sync.RWMutex
	cache map[string]*template.Template
}

func NewRenderer() (*Renderer, error) {
	return &Renderer{
		cache: make(map[string]*template.Template),
	}, nil
}

func (r *Renderer) Render(w http.ResponseWriter, pageFile string, data any) {
	t, err := r.get(pageFile)
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (r *Renderer) get(pageFile string) (*template.Template, error) {
	r.mu.RLock()
	if t, ok := r.cache[pageFile]; ok {
		r.mu.RUnlock()
		return t, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if t, ok := r.cache[pageFile]; ok {
		return t, nil
	}

	layout := filepath.Join("templates", "layout.html")
	page := filepath.Join("templates", pageFile)

	t, err := template.ParseFiles(layout, page)
	if err != nil {
		return nil, err
	}
	r.cache[pageFile] = t
	return t, nil
}
