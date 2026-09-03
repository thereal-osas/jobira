package validator

import "testing"

func TestRequired(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "normal value",
			value:    "Jobira",
			expected: true,
		},
		{
			name:     "leading and trailing spaces",
			value:    "  Jobira  ",
			expected: true,
		},
		{
			name:     "empty",
			value:    "",
			expected: false,
		},
		{
			name:     "spaces only",
			value:    "   ",
			expected: false,
		},
		{
			name:     "tabs and newlines only",
			value:    "\t\n ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Required(tt.value)

			if got != tt.expected {
				t.Fatalf(
					"Required(%q) = %v, expected %v",
					tt.value,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		min      int
		expected bool
	}{
		{
			name:     "greater than minimum",
			value:    "Jobira",
			min:      3,
			expected: true,
		},
		{
			name:     "exact minimum",
			value:    "Jobira",
			min:      6,
			expected: true,
		},
		{
			name:     "below minimum",
			value:    "Jobira",
			min:      7,
			expected: false,
		},
		{
			name:     "trims surrounding whitespace",
			value:    "  Jobira  ",
			min:      6,
			expected: true,
		},
		{
			name:     "whitespace does not count",
			value:    "  Jobira  ",
			min:      7,
			expected: false,
		},
		{
			name:     "empty with minimum one",
			value:    "",
			min:      1,
			expected: false,
		},
		{
			name:     "empty with minimum zero",
			value:    "",
			min:      0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MinLength(
				tt.value,
				tt.min,
			)

			if got != tt.expected {
				t.Fatalf(
					"MinLength(%q, %d) = %v, expected %v",
					tt.value,
					tt.min,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestIsEmail_Valid(t *testing.T) {
	tests := []string{
		"user@example.com",
		"cleaner@jobira.co.uk",
		"first.last@example.com",
		"user+tag@example.com",
		"USER@example.com",
	}

	for _, email := range tests {
		t.Run(email, func(t *testing.T) {
			if !IsEmail(email) {
				t.Fatalf(
					"expected %q to be valid",
					email,
				)
			}
		})
	}
}

func TestIsEmail_TrimsWhitespace(t *testing.T) {
	if !IsEmail("  user@example.com  ") {
		t.Fatal(
			"expected whitespace-trimmed email to be valid",
		)
	}
}

func TestIsEmail_Invalid(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"user",
		"user@",
		"@example.com",
		"user@example",
		"user@.com",
		"user@example.",
		"user example@example.com",
		"user@@example.com",
	}

	for _, email := range tests {
		t.Run(email, func(t *testing.T) {
			if IsEmail(email) {
				t.Fatalf(
					"expected %q to be invalid",
					email,
				)
			}
		})
	}
}

func TestIsEmail_RejectsDisplayName(t *testing.T) {
	value := "Jobira User <user@example.com>"

	if IsEmail(value) {
		t.Fatalf(
			"expected display-name address %q to be rejected",
			value,
		)
	}
}
