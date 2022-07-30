package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 5 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// nolint: errorlint,stylecheck
func run() error {
	rand.Seed(time.Now().UnixNano())
	log.Print("Starting...")

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	conf, err := ReadConfig()
	if err != nil {
		return fmt.Errorf("Failed to read config: %v", err)
	}

	// Init database
	db, err := OpenStorage(conf.DataFile, int64(conf.DataFileMaxSize))
	if err != nil {
		return fmt.Errorf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Init http server
	srv, err := NewServer(db, conf.Port, conf.TemplatesDir, conf.StaticDir)
	if err != nil {
		return fmt.Errorf("Failed to init server: %v", err)
	}

	// Wait for stop signal
	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Listening on %s", srv.Addr)
	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("Server failed with error: %v", err)
	}
	log.Print("Shutdown gracefully")
	return nil
}
