package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rodrigueghenda/jobira/internal/app"
	"github.com/rodrigueghenda/jobira/internal/config"
	"github.com/rodrigueghenda/jobira/internal/db"
)

func main() {
	cfg := config.Load()

	database, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatalf(
			"failed to connect to database: %v",
			err,
		)
	}
	defer database.Close()

	application, err := app.New(
		cfg,
		database,
	)
	if err != nil {
		log.Fatalf(
			"failed to initialize app: %v",
			err,
		)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           application.Router(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownSignal := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		shutdownSignal,
		os.Interrupt,
		syscall.SIGTERM,
	)

	serverErrors := make(
		chan error,
		1,
	)

	go func() {
		log.Printf(
			"server running on port %s",
			cfg.Port,
		)

		serverErrors <- server.ListenAndServe()
	}()

	select {
	case sig := <-shutdownSignal:
		log.Printf(
			"shutdown signal received: %s",
			sig,
		)

	case err := <-serverErrors:
		if err != nil &&
			!errors.Is(
				err,
				http.ErrServerClosed,
			) {
			log.Printf(
				"server failed: %v",
				err,
			)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(
		shutdownCtx,
	); err != nil {
		log.Printf(
			"graceful shutdown failed: %v",
			err,
		)

		if closeErr := server.Close(); closeErr != nil {
			log.Printf(
				"forced server close failed: %v",
				closeErr,
			)
		}
	}

	log.Println(
		"server stopped",
	)
}
