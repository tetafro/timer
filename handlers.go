package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	// idLength is a number of chars in timer id.
	idLength = 8
	// timeLayout is a layout for parsing time from client.
	timeLayout = "2006-01-02T15:04 -07:00"
)

// Timer represents a named point in time.
type Timer struct {
	Name     string `json:"name"`
	Deadline int64  `json:"deadline"`
}

// NewServer initializes new HTTP server with its handlers.
func NewServer(
	s *Storage,
	port int,
	tplDir string,
	staticDir string,
) (*http.Server, error) {
	// Parse templates
	indexTpl, err := template.ParseFiles(
		filepath.Join(tplDir, "base.html"),
		filepath.Join(tplDir, "index.html"),
	)
	if err != nil {
		return nil, fmt.Errorf("parse index template: %w", err)
	}
	timerTpl, err := template.ParseFiles(
		filepath.Join(tplDir, "base.html"),
		filepath.Join(tplDir, "timer.html"),
	)
	if err != nil {
		return nil, fmt.Errorf("parse timer template: %w", err)
	}

	// Init main router
	r := chi.NewRouter()
	r.Get("/", RootHandler(indexTpl))
	r.Post("/timers", CreateTimerHandler(s))
	r.Get("/timers/{id}", GetTimerHandler(s, timerTpl))

	// Serve static content
	fileHandler := http.FileServer(http.Dir(staticDir))
	r.Get("/static/*", http.StripPrefix("/static", fileHandler).ServeHTTP)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	return srv, nil
}

// RootHandler creates HTTP handler for the root directory.
func RootHandler(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, nil) // nolint: errcheck,gosec
	}
}

// GetTimerHandler creates HTTP handler for getting timers by id.
func GetTimerHandler(st *Storage, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		t, err := st.GetTimer(id)
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("Failed to get timer from storage %s: %v", id, err)
			internalServerError(w)
			return
		}

		// Render page
		tpl.Execute(w, t) // nolint: errcheck,gosec
	}
}

// CreateTimerHandler creates HTTP handler for creating new timers.
func CreateTimerHandler(st *Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse time
		t, err := parseTime(
			strings.TrimSpace(r.FormValue("time")),
			strings.TrimSpace(r.FormValue("timezone")),
		)
		if err != nil {
			log.Printf("Failed to parse time: %v", err)
			badRequest(w)
			return
		}

		id, err := st.SaveTimer(Timer{
			Name:     strings.TrimSpace(r.FormValue("name")),
			Deadline: t.Unix(),
		})
		if err != nil {
			log.Printf("Failed to save timer in storage: %v", err)
			internalServerError(w)
			return
		}

		// Redirect to timer's page
		http.Redirect(w, r,
			"/timers/"+id,
			http.StatusSeeOther)
	}
}

func parseTime(t, tz string) (time.Time, error) {
	return time.Parse(timeLayout, fmt.Sprintf("%s %s", t, tz)) // nolint: wrapcheck
}

func badRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(http.StatusText(http.StatusBadRequest))) // nolint: errcheck,gosec
}

func internalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError))) // nolint: errcheck,gosec
}
