package config

import "os"

type Config struct {
	Port 		string
	DatabaseURL string
	AppEnv 		string
	JWTSecret   string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	AppEnv := os.Getenv("APP_ENV")
	if AppEnv == "" {
		AppEnv = "development"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}

	return Config{
		Port: 	     port,
		DatabaseURL: dbURL,
		AppEnv: 	 AppEnv,	 	
		JWTSecret:   jwtSecret,
	}
}