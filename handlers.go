package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

// idLength is a number of chars in timer id.
const idLength = 8

// NewServer initializes new HTTP server with its handlers.
func NewServer(
	s *Storage,
	port int,
	basePath string,
	tplDir string,
	staticDir string,
	reqlimCount int,
	reqlimWindow time.Duration,
) (*http.Server, error) {
	h := Handler{storage: s, basePath: basePath}

	// Parse templates
	var err error
	h.templates.index, err = template.ParseFiles(
		filepath.Join(tplDir, "base.html"),
		filepath.Join(tplDir, "index.html"),
	)
	if err != nil {
		return nil, fmt.Errorf("parse index template: %w", err)
	}
	h.templates.timer, err = template.ParseFiles(
		filepath.Join(tplDir, "base.html"),
		filepath.Join(tplDir, "timer.html"),
	)
	if err != nil {
		return nil, fmt.Errorf("parse timer template: %w", err)
	}
	h.templates.error, err = template.ParseFiles(
		filepath.Join(tplDir, "base.html"),
		filepath.Join(tplDir, "error.html"),
	)
	if err != nil {
		return nil, fmt.Errorf("parse error template: %w", err)
	}

	// Init main router
	r := chi.NewRouter()

	// Setup rate limit by ip/endpoint pair
	if reqlimCount > 0 {
		r.Use(httprate.Limit(
			reqlimCount, reqlimWindow,
			httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		))
	}

	// Setup routes
	r.Get("/", h.Index)
	r.Post("/timers", h.CreateTimer)
	r.Get("/timers/{id}", h.GetTimer)

	// Serve static content
	fileHandler := http.FileServer(http.Dir(staticDir))
	r.Get("/static/*", http.StripPrefix("/static", fileHandler).ServeHTTP)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return srv, nil
}

// Handler provides a set of HTTP handlers for all server routes.
type Handler struct {
	storage   *Storage
	basePath  string
	templates struct {
		index *template.Template
		timer *template.Template
		error *template.Template
	}
}

// Timer represents a point in time.
type Timer struct {
	Deadline int64 `json:"deadline"`
	Created  int64 `json:"created"`
}

// Context is data shared across all templates.
type Context struct {
	BasePath string
}

// ErrorPage contains data for the page that shows an error.
type ErrorPage struct {
	Context
	Error string
}

// TimerPage contains data for the page that shows a timer.
type TimerPage struct {
	Context
	Deadline    int64
	WithMinutes bool
	WithHours   bool
	WithDays    bool
}

// Index handles HTTP requests for the root directory.
func (h *Handler) Index(w http.ResponseWriter, _ *http.Request) {
	data := Context{BasePath: h.basePath}
	w.WriteHeader(http.StatusOK)
	h.templates.index.Execute(w, data) //nolint:errcheck,gosec
}

// GetTimer handles HTTP requests for getting timers by id.
func (h *Handler) GetTimer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	t, err := h.storage.GetTimer(id)
	if errors.Is(err, ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("Failed to get timer from storage %s: %v", id, err)
		h.internalServerError(w)
		return
	}

	data := TimerPage{
		Deadline:    t.Deadline,
		WithMinutes: t.Deadline-t.Created >= 60,
		WithHours:   t.Deadline-t.Created >= 60*60,
		WithDays:    t.Deadline-t.Created >= 60*60*24,
	}

	// Render page
	h.templates.timer.Execute(w, data) //nolint:errcheck,gosec
}

// CreateTimer handles HTTP requests for creating new timers.
func (h *Handler) CreateTimer(w http.ResponseWriter, r *http.Request) {
	t, err := time.Parse(time.RFC3339, r.FormValue("deadline"))
	if err != nil {
		h.badRequest(w, "Invalid time format")
		return
	}

	id, err := h.storage.SaveTimer(Timer{
		Deadline: t.Unix(),
		Created:  time.Now().Unix(),
	})
	if err != nil {
		log.Printf("Failed to save timer in storage: %v", err)
		h.internalServerError(w)
		return
	}

	// Redirect to timer's page
	http.Redirect(w, r,
		"/timers/"+id,
		http.StatusSeeOther)
}

func (h *Handler) internalServerError(w http.ResponseWriter) {
	data := ErrorPage{
		Context: Context{BasePath: h.basePath},
		Error:   "Internal server error",
	}
	w.WriteHeader(http.StatusInternalServerError)
	h.templates.error.Execute(w, data) //nolint:errcheck,gosec
}

func (h *Handler) badRequest(w http.ResponseWriter, msg string) {
	data := ErrorPage{
		Context: Context{BasePath: h.basePath},
		Error:   msg,
	}
	w.WriteHeader(http.StatusBadRequest)
	h.templates.error.Execute(w, data) //nolint:errcheck,gosec
}
