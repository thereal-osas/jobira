package password

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHash_Success(t *testing.T) {
	raw := "StrongPassword123!"

	hashed, err := Hash(raw)

	if err != nil {
		t.Fatalf(
			"unexpected hash error: %v",
			err,
		)
	}

	if hashed == "" {
		t.Fatal("expected password hash")
	}

	if hashed == raw {
		t.Fatal(
			"hashed password must not equal raw password",
		)
	}

	if !strings.HasPrefix(hashed, "$2") {
		t.Fatalf(
			"expected bcrypt hash, got %q",
			hashed,
		)
	}
}

func TestHash_ProducesDifferentHashes(
	t *testing.T,
) {
	raw := "StrongPassword123!"

	first, err := Hash(raw)
	if err != nil {
		t.Fatalf(
			"unexpected first hash error: %v",
			err,
		)
	}

	second, err := Hash(raw)
	if err != nil {
		t.Fatalf(
			"unexpected second hash error: %v",
			err,
		)
	}

	if first == second {
		t.Fatal(
			"expected bcrypt salt to produce different hashes",
		)
	}

	if err := Compare(first, raw); err != nil {
		t.Fatalf(
			"first hash should verify: %v",
			err,
		)
	}

	if err := Compare(second, raw); err != nil {
		t.Fatalf(
			"second hash should verify: %v",
			err,
		)
	}
}

func TestCompare_CorrectPassword(t *testing.T) {
	raw := "MySecurePassword!"

	hashed, err := Hash(raw)
	if err != nil {
		t.Fatalf(
			"unexpected hash error: %v",
			err,
		)
	}

	err = Compare(hashed, raw)

	if err != nil {
		t.Fatalf(
			"expected password match, got %v",
			err,
		)
	}
}

func TestCompare_IncorrectPassword(t *testing.T) {
	hashed, err := Hash("correct-password")
	if err != nil {
		t.Fatalf(
			"unexpected hash error: %v",
			err,
		)
	}

	err = Compare(
		hashed,
		"wrong-password",
	)

	if !errors.Is(
		err,
		bcrypt.ErrMismatchedHashAndPassword,
	) {
		t.Fatalf(
			"expected ErrMismatchedHashAndPassword, got %v",
			err,
		)
	}
}

func TestCompare_InvalidHash(t *testing.T) {
	err := Compare(
		"not-a-valid-bcrypt-hash",
		"password",
	)

	if err == nil {
		t.Fatal("expected invalid hash error")
	}
}

func TestCompare_EmptyHash(t *testing.T) {
	err := Compare("", "password")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCompare_EmptyPasswordDoesNotMatch(
	t *testing.T,
) {
	hashed, err := Hash("password")
	if err != nil {
		t.Fatalf(
			"unexpected hash error: %v",
			err,
		)
	}

	err = Compare(hashed, "")

	if err == nil {
		t.Fatal(
			"expected empty password not to match",
		)
	}
}

func TestHashAndCompare_UnicodePassword(
	t *testing.T,
) {
	raw := "Jøbira-安全-123!"

	hashed, err := Hash(raw)
	if err != nil {
		t.Fatalf(
			"unexpected hash error: %v",
			err,
		)
	}

	if err := Compare(hashed, raw); err != nil {
		t.Fatalf(
			"expected unicode password to match: %v",
			err,
		)
	}
}

func TestHash_PasswordTooLong(t *testing.T) {
	raw := strings.Repeat("a", 73)

	_, err := Hash(raw)

	if !errors.Is(
		err,
		bcrypt.ErrPasswordTooLong,
	) {
		t.Fatalf(
			"expected ErrPasswordTooLong, got %v",
			err,
		)
	}
}
