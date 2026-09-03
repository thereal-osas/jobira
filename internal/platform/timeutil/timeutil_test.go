package timeutil

import (
	"testing"
	"time"
)

func TestNowTC_ReturnsUTC(t *testing.T) {
	before := time.Now().UTC()

	got := NowTC()

	after := time.Now().UTC()

	if got.Location() != time.UTC {
		t.Fatalf(
			"expected UTC location, got %v",
			got.Location(),
		)
	}

	if got.Before(before) {
		t.Fatalf(
			"expected time >= %v, got %v",
			before,
			got,
		)
	}

	if got.After(after) {
		t.Fatalf(
			"expected time <= %v, got %v",
			after,
			got,
		)
	}
}

func TestFormatISO(t *testing.T) {
	input := time.Date(
		2026,
		time.September,
		1,
		14,
		30,
		45,
		0,
		time.UTC,
	)

	got := FormatISO(input)

	expected := "2026-09-01T14:30:45Z"

	if got != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			got,
		)
	}
}

func TestFormatISO_ConvertsToUTC(t *testing.T) {
	location := time.FixedZone(
		"BST",
		60*60,
	)

	input := time.Date(
		2026,
		time.September,
		1,
		15,
		30,
		45,
		0,
		location,
	)

	got := FormatISO(input)

	expected := "2026-09-01T14:30:45Z"

	if got != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			got,
		)
	}
}

func TestFormatISO_RemovesSubSecondPrecision(
	t *testing.T,
) {
	input := time.Date(
		2026,
		time.September,
		1,
		14,
		30,
		45,
		987654321,
		time.UTC,
	)

	got := FormatISO(input)

	expected := "2026-09-01T14:30:45Z"

	if got != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			got,
		)
	}
}

func TestFormatISO_ZeroTime(t *testing.T) {
	got := FormatISO(time.Time{})

	expected := "0001-01-01T00:00:00Z"

	if got != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			got,
		)
	}
}
