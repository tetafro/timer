package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/fsnotify/fsnotify"
)

const (
	openDatabaseTimeout = 1 * time.Second
	bucketName          = "timers"
)

// List of public errors returned by storage.
var (
	ErrNotFound = errors.New("timer not found")
	ErrIsFull   = errors.New("storage is full")
)

// Storage is responsible for storing timers.
type Storage struct {
	db      *bolt.DB
	watcher *fsnotify.Watcher
	limit   int64
	full    bool
	mx      sync.Mutex
}

// OpenStorage opens or creates new storage database with a default bucket.
func OpenStorage(datafile string, limit int64) (*Storage, error) {
	db, err := bolt.Open(datafile, 0o600, &bolt.Options{Timeout: openDatabaseTimeout})
	if err != nil {
		return nil, fmt.Errorf("open database file: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err //nolint:wrapcheck
	})
	if err != nil {
		return nil, fmt.Errorf("init database bucket: %w", err)
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create file watcher: %w", err)
	}
	if err := w.Add(datafile); err != nil {
		return nil, fmt.Errorf("add file for watching: %w", err)
	}

	s := &Storage{
		db:      db,
		watcher: w,
		limit:   limit,
	}

	go s.watchDBSize(datafile)

	return s, nil
}

// GetTimer gets timer by id from storage.
func (s *Storage) GetTimer(id string) (Timer, error) {
	var data []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		data = tx.Bucket([]byte(bucketName)).Get([]byte(id))
		return nil
	})
	if err != nil {
		return Timer{}, fmt.Errorf("get timer from database: %w", err)
	}
	if data == nil {
		return Timer{}, ErrNotFound
	}

	var t Timer
	if err := json.Unmarshal(data, &t); err != nil {
		return Timer{}, fmt.Errorf("unmarshal database data: %w", err)
	}
	return t, nil
}

// SaveTimer saves timer in storage and returns its id.
func (s *Storage) SaveTimer(t Timer) (string, error) {
	if s.isFull() {
		return "", ErrIsFull
	}

	data, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("marshal database data: %w", err)
	}

	id := generateID()
	err = s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).Put([]byte(id), data)
	})
	if err != nil {
		return "", fmt.Errorf("save timer in database: %w", err)
	}

	return id, nil
}

// Close closes db properly.
func (s *Storage) Close() {
	if err := s.db.Close(); err != nil {
		log.Printf("Failed to properly close database: %v", err)
	}
	if err := s.watcher.Close(); err != nil {
		log.Printf("Failed to properly close database file watcher: %v", err)
	}
}

// watchDBSize sets up a file watcher. When it gets a write event, the database
// file's size is checked, and if it exceeds the limit, `full` flag is set to
// prevent any further writes.
//
//nolint:cyclop
func (s *Storage) watchDBSize(datafile string) {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				stat, err := os.Stat(datafile)
				if err != nil {
					log.Printf("Failed to read file stat: %v", err)
					continue
				}
				switch {
				case stat.Size() >= s.limit && !s.isFull():
					s.setFull()
					log.Print("Database file size limit is reached")
				case stat.Size() < s.limit && s.isFull():
					s.unsetFull()
					log.Print("Database file size is ok now")
				}
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Database file watch error: %v", err)
		}
	}
}

func (s *Storage) isFull() bool {
	s.mx.Lock()
	defer s.mx.Unlock()

	return s.full
}

func (s *Storage) setFull() {
	s.mx.Lock()
	s.full = true
	s.mx.Unlock()
}

func (s *Storage) unsetFull() {
	s.mx.Lock()
	s.full = false
	s.mx.Unlock()
}

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, idLength)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}
	return string(id)
}
