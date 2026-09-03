package env

import (
	"os"
	"testing"
)

func TestGet_ReturnsEnvironmentValue(t *testing.T) {
	t.Setenv(
		"JOBIRA_TEST_ENV",
		"configured-value",
	)

	got := Get(
		"JOBIRA_TEST_ENV",
		"fallback-value",
	)

	if got != "configured-value" {
		t.Fatalf(
			"expected configured-value, got %q",
			got,
		)
	}
}

func TestGet_ReturnsFallbackWhenMissing(
	t *testing.T,
) {
	const key = "JOBIRA_TEST_MISSING_ENV"

	_ = os.Unsetenv(key)

	got := Get(
		key,
		"fallback-value",
	)

	if got != "fallback-value" {
		t.Fatalf(
			"expected fallback-value, got %q",
			got,
		)
	}
}

func TestGet_ReturnsFallbackWhenEmpty(
	t *testing.T,
) {
	t.Setenv(
		"JOBIRA_TEST_EMPTY_ENV",
		"",
	)

	got := Get(
		"JOBIRA_TEST_EMPTY_ENV",
		"fallback",
	)

	if got != "fallback" {
		t.Fatalf(
			"expected fallback, got %q",
			got,
		)
	}
}

func TestRequired_ReturnsEnvironmentValue(
	t *testing.T,
) {
	t.Setenv(
		"JOBIRA_REQUIRED_ENV",
		"required-value",
	)

	got := Required(
		"JOBIRA_REQUIRED_ENV",
	)

	if got != "required-value" {
		t.Fatalf(
			"expected required-value, got %q",
			got,
		)
	}
}

func TestRequired_PanicsWhenMissing(
	t *testing.T,
) {
	const key = "JOBIRA_REQUIRED_MISSING"

	_ = os.Unsetenv(key)

	defer func() {
		rec := recover()

		if rec == nil {
			t.Fatal(
				"expected Required to panic",
			)
		}

		expected := key + " is required"

		if rec != expected {
			t.Fatalf(
				"expected panic %q, got %v",
				expected,
				rec,
			)
		}
	}()

	Required(key)
}

func TestRequired_PanicsWhenEmpty(
	t *testing.T,
) {
	const key = "JOBIRA_REQUIRED_EMPTY"

	t.Setenv(key, "")

	defer func() {
		rec := recover()

		if rec == nil {
			t.Fatal(
				"expected Required to panic",
			)
		}

		expected := key + " is required"

		if rec != expected {
			t.Fatalf(
				"expected panic %q, got %v",
				expected,
				rec,
			)
		}
	}()

	Required(key)
}
