package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	badger "github.com/dgraph-io/badger/v3"
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
	db *badger.DB,
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
	r.Post("/timer", CreateTimerHandler(db))
	r.Get("/timer/{id}", GetTimerHandler(db, timerTpl))

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
func GetTimerHandler(db *badger.DB, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		// Get timer from db
		var timer Timer
		err := db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(id))
			if err != nil {
				return fmt.Errorf("get item: %w", err)
			}
			val, err := item.ValueCopy(nil)
			if err != nil {
				return fmt.Errorf("read value: %w", err)
			}
			if err := json.Unmarshal(val, &timer); err != nil {
				return fmt.Errorf("unmarshal db data: %w", err)
			}
			return nil
		})
		if errors.Is(err, badger.ErrKeyNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("Failed to get timer from db %s: %v", id, err)
			internalServerError(w)
			return
		}

		// Render page
		tpl.Execute(w, timer) // nolint: errcheck,gosec
	}
}

// CreateTimerHandler creates HTTP handler for creating new timers.
func CreateTimerHandler(db *badger.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse time
		t, err := parseTime(
			r.FormValue("time"),
			r.FormValue("timezone"),
		)
		if err != nil {
			log.Printf("Failed to parse time: %v", err)
			badRequest(w)
			return
		}
		timer := Timer{
			Name:     r.FormValue("name"),
			Deadline: t.Unix(),
		}

		// Save timer to db
		id := generateID()
		data, err := json.Marshal(timer)
		if err != nil {
			log.Printf("Failed to marshal db data: %v", err)
			internalServerError(w)
			return
		}
		err = db.Update(func(txn *badger.Txn) error {
			return txn.Set(id, data) // nolint: wrapcheck
		})
		if err != nil {
			log.Printf("Failed to save timer in db: %v", err)
			internalServerError(w)
			return
		}

		// Redirect to timer's page
		http.Redirect(w, r,
			"/timer/"+string(id),
			http.StatusSeeOther)
	}
}

func generateID() []byte {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, idLength)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}
	return id
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
