package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var direction string

	flag.StringVar(
		&direction,
		"direction",
		"up",
		"migration direction: up or down",
	)

	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatalf(
			"create migrator: %v",
			err,
		)
	}

	defer func() {
		sourceErr, databaseErr := m.Close()

		if sourceErr != nil {
			log.Printf(
				"close migration source: %v",
				sourceErr,
			)
		}

		if databaseErr != nil {
			log.Printf(
				"close migration database: %v",
				databaseErr,
			)
		}
	}()

	switch direction {
	case "up":
		err = m.Up()

	case "down":
		err = m.Steps(-1)

	default:
		log.Fatalf(
			"unsupported migration direction %q; use up or down",
			direction,
		)
	}

	if err != nil {
		if errors.Is(
			err,
			migrate.ErrNoChange,
		) {
			fmt.Println("database already up to date")
			return
		}

		log.Fatalf(
			"migration failed: %v",
			err,
		)
	}

	fmt.Printf(
		"migration %s completed successfully\n",
		direction,
	)
}
