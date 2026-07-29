package main

import (
	"log"
	"net/http"
	"github.com/rodrigueghenda/jobira/internal/app"
	"github.com/rodrigueghenda/jobira/internal/config"
	"github.com/rodrigueghenda/jobira/internal/db"
	
)

func main() {
	cfg := config.Load()

	database, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect to databse: %v", err)
	}

	application, err := app.New(cfg, database)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	log.Printf("server running on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, application.Router()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
 

}