package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
	AppEnv      string
	JWTSecret   string

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	StripeSecretKey       string
	StripeWebhookSecret   string
	StripeSuccessURL      string
	StripeCancelURL       string
	StripePortalReturnURL string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appEnv := strings.ToLower(
		strings.TrimSpace(
			os.Getenv("APP_ENV"),
		),
	)
	if appEnv == "" {
		appEnv = "development"
	}

	dbURL := strings.TrimSpace(
		os.Getenv("DATABASE_URL"),
	)

	jwtSecret := strings.TrimSpace(
		os.Getenv("JWT_SECRET"),
	)

	if jwtSecret == "" {
		if appEnv == "production" {
			panic(
				"JWT_SECRET is required in production",
			)
		}

		jwtSecret = "change-me-in-development"
	}

	smtpHost := strings.TrimSpace(
		os.Getenv("SMTP_HOST"),
	)
	smtpPort := strings.TrimSpace(
		os.Getenv("SMTP_PORT"),
	)
	smtpUsername := strings.TrimSpace(
		os.Getenv("SMTP_USERNAME"),
	)
	smtpPassword := strings.TrimSpace(
		os.Getenv("SMTP_PASSWORD"),
	)
	smtpFrom := strings.TrimSpace(
		os.Getenv("SMTP_FROM"),
	)

	stripeSecretKey := strings.TrimSpace(
		os.Getenv("STRIPE_SECRET_KEY"),
	)

	stripeWebhookSecret := strings.TrimSpace(
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
	)

	stripeSuccessURL := strings.TrimSpace(
		os.Getenv("STRIPE_SUCCESS_URL"),
	)

	stripeCancelURL := strings.TrimSpace(
		os.Getenv("STRIPE_CANCEL_URL"),
	)

	stripePortalReturnURL := strings.TrimSpace(
		os.Getenv("STRIPE_PORTAL_RETURN_URL"),
	)

	if appEnv == "production" {
		requiredSMTP := map[string]string{
			"SMTP_HOST": smtpHost,
			"SMTP_PORT": smtpPort,
			"SMTP_FROM": smtpFrom,
		}

		requiredStripe := map[string]string{
			"STRIPE_SECRET_KEY":        stripeSecretKey,
			"STRIPE_WEBHOOK_SECRET":    stripeWebhookSecret,
			"STRIPE_SUCCESS_URL":       stripeSuccessURL,
			"STRIPE_CANCEL_URL":        stripeCancelURL,
			"STRIPE_PORTAL_RETURN_URL": stripePortalReturnURL,
		}

		for name, value := range requiredStripe {
			if value == "" {
				panic(
					fmt.Sprintf(
						"%s is required in production",
						name,
					),
				)
			}
		}

		for name, value := range requiredSMTP {
			if value == "" {
				panic(
					fmt.Sprintf(
						"%s is required in production",
						name,
					),
				)
			}
		}
	}

	return Config{
		Port:                  port,
		DatabaseURL:           dbURL,
		AppEnv:                appEnv,
		JWTSecret:             jwtSecret,
		SMTPHost:              smtpHost,
		SMTPPort:              smtpPort,
		SMTPUsername:          smtpUsername,
		SMTPPassword:          smtpPassword,
		SMTPFrom:              smtpFrom,
		StripeSecretKey:       stripeSecretKey,
		StripeWebhookSecret:   stripeWebhookSecret,
		StripeSuccessURL:      stripeSuccessURL,
		StripeCancelURL:       stripeCancelURL,
		StripePortalReturnURL: stripePortalReturnURL,
	}
}
