package app

import (
	"database/sql"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/config"
	"github.com/rodrigueghenda/jobira/internal/transport/http/middleware"
)

type App struct {
	Cfg config.Config
	DB  *sql.DB
	R 	*chi.Mux
}

func New(cfg config.Config, db *sql.DB) (*App, error) {
	r := chi.NewRouter()
 
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recover)
	r.Use(rateLimiter.Middleware)

	app := &App{
		Cfg: cfg,
		DB: db,
		R: r,
	}

	app.registerRoutes()

	return app, nil
}   