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
	tplDir string,
	staticDir string,
	reqlimCount int,
	reqlimWindow time.Duration,
) (*http.Server, error) {
	h := Handler{storage: s}

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
	templates struct {
		index *template.Template
		timer *template.Template
	}
}

// Timer represents a point in time.
type Timer struct {
	Deadline int64 `json:"deadline"`
	Created  int64 `json:"created"`
}

// TimerPage contains data for the page that shows a timer.
type TimerPage struct {
	Deadline    int64
	WithMinutes bool
	WithHours   bool
	WithDays    bool
}

// Index handles HTTP requests for the root directory.
func (h *Handler) Index(w http.ResponseWriter, _ *http.Request) {
	h.index(w, http.StatusOK, "")
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

func (h *Handler) index(w http.ResponseWriter, code int, err string) {
	data := struct {
		Error string
	}{
		Error: err,
	}
	w.WriteHeader(code)
	h.templates.index.Execute(w, data) //nolint:errcheck,gosec
}

func (h *Handler) internalServerError(w http.ResponseWriter) {
	h.index(w, http.StatusInternalServerError, "Internal server error")
}

func (h *Handler) badRequest(w http.ResponseWriter, msg string) {
	h.index(w, http.StatusBadRequest, msg)
}
