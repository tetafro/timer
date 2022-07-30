package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/boltdb/bolt"
)

const (
	openDatabaseTimeout = 1 * time.Second
	bucketName          = "timers"
)

// List of public errors returned by storage.
var (
	ErrNotFound = errors.New("timer not found")
)

// Storage is responsible for storing timers.
type Storage struct {
	db *bolt.DB
}

// OpenStorage opens or creates new storage database with a default bucket.
func OpenStorage(datafile string) (*Storage, error) {
	db, err := bolt.Open(datafile, 0o600, &bolt.Options{Timeout: openDatabaseTimeout})
	if err != nil {
		return nil, fmt.Errorf("open database file: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err // nolint: wrapcheck
	})
	if err != nil {
		return nil, fmt.Errorf("init database bucket: %w", err)
	}

	s := &Storage{
		db: db,
	}

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
	data, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("marshal database data: %w", err)
	}

	id := generateID()
	err = s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).Put([]byte(id), data) // nolint: wrapcheck
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
}

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, idLength)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}
	return string(id)
}
