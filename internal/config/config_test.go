package config

import (
	"os"
	"testing"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"PORT",
		"DATABASE_URL",
		"APP_ENV",
		"JWT_SECRET",
	}

	for _, key := range keys {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf(
				"failed to unset %s: %v",
				key,
				err,
			)
		}
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearConfigEnv(t)

	cfg := Load()

	if cfg.Port != "8080" {
		t.Fatalf(
			"expected port 8080, got %q",
			cfg.Port,
		)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf(
			"expected development, got %q",
			cfg.AppEnv,
		)
	}

	if cfg.DatabaseURL != "" {
		t.Fatalf(
			"expected empty database URL, got %q",
			cfg.DatabaseURL,
		)
	}

	if cfg.JWTSecret != "change-me-in-development" {
		t.Fatalf(
			"unexpected JWT secret %q",
			cfg.JWTSecret,
		)
	}
}

func TestLoad_EnvironmentValues(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv(
		"DATABASE_URL",
		"postgres://localhost/jobira",
	)
	t.Setenv("APP_ENV", "test")
	t.Setenv(
		"JWT_SECRET",
		"super-secret-key",
	)

	cfg := Load()

	if cfg.Port != "9090" {
		t.Fatalf(
			"expected 9090, got %q",
			cfg.Port,
		)
	}

	if cfg.DatabaseURL != "postgres://localhost/jobira" {
		t.Fatalf(
			"unexpected database URL %q",
			cfg.DatabaseURL,
		)
	}

	if cfg.AppEnv != "test" {
		t.Fatalf(
			"expected test, got %q",
			cfg.AppEnv,
		)
	}

	if cfg.JWTSecret != "super-secret-key" {
		t.Fatalf(
			"unexpected JWT secret %q",
			cfg.JWTSecret,
		)
	}
}

func TestLoad_ProductionWithJWTSecret(
	t *testing.T,
) {
	clearConfigEnv(t)

	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "production-secret")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_FROM", "noreply@jobira.co.uk")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test_123")
	t.Setenv("STRIPE_SUCCESS_URL", "https://jobira.co.uk/billing/success")
	t.Setenv("STRIPE_CANCEL_URL", "https://jobira.co.uk/billing/cancel")
	t.Setenv("STRIPE_PORTAL_RETURN_URL", "https://jobira.co.uk/account/billing")

	cfg := Load()

	if cfg.AppEnv != "production" {
		t.Fatalf(
			"expected production, got %q",
			cfg.AppEnv,
		)
	}

	if cfg.JWTSecret != "production-secret" {
		t.Fatalf(
			"unexpected JWT secret %q",
			cfg.JWTSecret,
		)
	}
}

func TestLoad_ProductionWithoutSMTPHost_Panics(
	t *testing.T,
) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "production-secret")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_FROM", "noreply@jobira.co.uk")

	defer func() {
		if recover() == nil {
			t.Fatal(
				"expected panic when SMTP_HOST is missing",
			)
		}
	}()

	Load()
}

func TestLoad_ProductionWithoutSMTPPort_Panics(
	t *testing.T,
) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "production-secret")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_FROM", "noreply@jobira.co.uk")

	defer func() {
		if recover() == nil {
			t.Fatal(
				"expected panic when SMTP_PORT is missing",
			)
		}
	}()

	Load()
}

func TestLoad_ProductionWithoutSMTPFrom_Panics(
	t *testing.T,
) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "production-secret")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")

	defer func() {
		if recover() == nil {
			t.Fatal(
				"expected panic when SMTP_FROM is missing",
			)
		}
	}()

	Load()
}

func TestLoad_ProductionWithoutJWTSecretPanics(
	t *testing.T,
) {
	clearConfigEnv(t)

	t.Setenv("APP_ENV", "production")

	defer func() {
		rec := recover()

		if rec == nil {
			t.Fatal(
				"expected production config to panic",
			)
		}

		expected :=
			"JWT_SECRET is required in production"

		if rec != expected {
			t.Fatalf(
				"expected panic %q, got %v",
				expected,
				rec,
			)
		}
	}()

	Load()
}

func TestLoad_ProductionWithoutRequiredStripeConfig_Panics(
	t *testing.T,
) {
	requiredStripe := map[string]string{
		"STRIPE_SECRET_KEY":        "sk_test_123",
		"STRIPE_WEBHOOK_SECRET":    "whsec_test_123",
		"STRIPE_SUCCESS_URL":       "https://jobira.co.uk/billing/success",
		"STRIPE_CANCEL_URL":        "https://jobira.co.uk/billing/cancel",
		"STRIPE_PORTAL_RETURN_URL": "https://jobira.co.uk/account/billing",
	}

	for missing := range requiredStripe {
		t.Run(missing, func(t *testing.T) {
			t.Setenv("APP_ENV", "production")
			t.Setenv("JWT_SECRET", "production-secret")

			t.Setenv("SMTP_HOST", "smtp.example.com")
			t.Setenv("SMTP_PORT", "587")
			t.Setenv("SMTP_FROM", "noreply@jobira.co.uk")

			for name, value := range requiredStripe {
				if name == missing {
					t.Setenv(name, "")
					continue
				}

				t.Setenv(name, value)
			}

			defer func() {
				if recover() == nil {
					t.Fatalf(
						"expected panic when %s is missing",
						missing,
					)
				}
			}()

			Load()
		})
	}
}

func TestLoad_CustomPortOnly(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("PORT", "3000")

	cfg := Load()

	if cfg.Port != "3000" {
		t.Fatalf(
			"expected 3000, got %q",
			cfg.Port,
		)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf(
			"expected development, got %q",
			cfg.AppEnv,
		)
	}
}

func TestLoad_DatabaseURL(t *testing.T) {
	clearConfigEnv(t)

	expected :=
		"postgres://user:pass@localhost:5433/jobira"

	t.Setenv(
		"DATABASE_URL",
		expected,
	)

	cfg := Load()

	if cfg.DatabaseURL != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			cfg.DatabaseURL,
		)
	}
}
