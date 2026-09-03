package jwt

import (
	"strconv"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestNewIssuer(t *testing.T) {
	issuer := NewIssuer("test-secret")

	if issuer == nil {
		t.Fatal("expected issuer")
	}

	if string(issuer.secret) != "test-secret" {
		t.Fatalf(
			"expected test-secret, got %q",
			string(issuer.secret),
		)
	}
}

func TestIssuer_GenerateAndParse(t *testing.T) {
	issuer := NewIssuer("test-secret")

	tokenString, err := issuer.Generate(
		42,
		"cleaner@example.com",
		"cleaner",
	)

	if err != nil {
		t.Fatalf(
			"unexpected generate error: %v",
			err,
		)
	}

	if tokenString == "" {
		t.Fatal("expected token")
	}

	claims, err := issuer.Parse(tokenString)
	if err != nil {
		t.Fatalf(
			"unexpected parse error: %v",
			err,
		)
	}

	if claims.UserID != 42 {
		t.Fatalf(
			"expected user id 42, got %d",
			claims.UserID,
		)
	}

	if claims.Email != "cleaner@example.com" {
		t.Fatalf(
			"unexpected email %q",
			claims.Email,
		)
	}

	if claims.Role != "cleaner" {
		t.Fatalf(
			"expected cleaner role, got %q",
			claims.Role,
		)
	}

	if claims.Subject != "42" {
		t.Fatalf(
			"expected subject 42, got %q",
			claims.Subject,
		)
	}

	if claims.IssuedAt == nil {
		t.Fatal("expected issued_at")
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expected expires_at")
	}

	expectedExpiry := claims.IssuedAt.Time.Add(
		24 * time.Hour,
	)

	diff := claims.ExpiresAt.Time.Sub(
		expectedExpiry,
	)

	if diff < -time.Second || diff > time.Second {
		t.Fatalf(
			"expected expiry about 24 hours after issue, got %v",
			claims.ExpiresAt.Time.Sub(
				claims.IssuedAt.Time,
			),
		)
	}
}

func TestIssuer_GenerateSubjectMatchesUserID(
	t *testing.T,
) {
	issuer := NewIssuer("secret")

	userID := uint(12345)

	tokenString, err := issuer.Generate(
		userID,
		"user@example.com",
		"client",
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	claims, err := issuer.Parse(tokenString)
	if err != nil {
		t.Fatalf(
			"unexpected parse error: %v",
			err,
		)
	}

	expected := strconv.FormatUint(
		uint64(userID),
		10,
	)

	if claims.Subject != expected {
		t.Fatalf(
			"expected subject %q, got %q",
			expected,
			claims.Subject,
		)
	}
}

func TestIssuer_Parse_WrongSecret(t *testing.T) {
	issuer := NewIssuer("correct-secret")

	tokenString, err := issuer.Generate(
		1,
		"user@example.com",
		"client",
	)

	if err != nil {
		t.Fatalf(
			"unexpected generate error: %v",
			err,
		)
	}

	wrongIssuer := NewIssuer("wrong-secret")

	claims, err := wrongIssuer.Parse(tokenString)

	if err == nil {
		t.Fatal("expected parse error")
	}

	if claims != nil {
		t.Fatal("expected nil claims")
	}
}

func TestIssuer_Parse_MalformedToken(t *testing.T) {
	issuer := NewIssuer("secret")

	claims, err := issuer.Parse(
		"definitely-not-a-token",
	)

	if err == nil {
		t.Fatal("expected parse error")
	}

	if claims != nil {
		t.Fatal("expected nil claims")
	}
}

func TestIssuer_Parse_ExpiredToken(t *testing.T) {
	issuer := NewIssuer("secret")

	now := time.Now()

	claims := Claims{
		UserID: 1,
		Email:  "user@example.com",
		Role:   "client",
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject: "1",
			IssuedAt: jwtlib.NewNumericDate(
				now.Add(-2 * time.Hour),
			),
			ExpiresAt: jwtlib.NewNumericDate(
				now.Add(-time.Hour),
			),
		},
	}

	token := jwtlib.NewWithClaims(
		jwtlib.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(
		[]byte("secret"),
	)
	if err != nil {
		t.Fatalf(
			"failed to sign token: %v",
			err,
		)
	}

	parsedClaims, err := issuer.Parse(tokenString)

	if err == nil {
		t.Fatal("expected expired token error")
	}

	if parsedClaims != nil {
		t.Fatal("expected nil claims")
	}
}

func TestIssuer_Parse_RejectsHS384(t *testing.T) {
	issuer := NewIssuer("secret")

	claims := Claims{
		UserID: 1,
		Email:  "user@example.com",
		Role:   "client",
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject: "1",
			ExpiresAt: jwtlib.NewNumericDate(
				time.Now().Add(time.Hour),
			),
		},
	}

	token := jwtlib.NewWithClaims(
		jwtlib.SigningMethodHS384,
		claims,
	)

	tokenString, err := token.SignedString(
		[]byte("secret"),
	)
	if err != nil {
		t.Fatalf(
			"failed to sign token: %v",
			err,
		)
	}

	parsedClaims, err := issuer.Parse(tokenString)

	if err == nil {
		t.Fatal(
			"expected unexpected signing method error",
		)
	}

	if parsedClaims != nil {
		t.Fatal("expected nil claims")
	}
}

func TestIssuer_Parse_RejectsUnsignedToken(
	t *testing.T,
) {
	issuer := NewIssuer("secret")

	claims := Claims{
		UserID: 1,
		Email:  "user@example.com",
		Role:   "client",
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(
				time.Now().Add(time.Hour),
			),
		},
	}

	token := jwtlib.NewWithClaims(
		jwtlib.SigningMethodNone,
		claims,
	)

	tokenString, err := token.SignedString(
		jwtlib.UnsafeAllowNoneSignatureType,
	)
	if err != nil {
		t.Fatalf(
			"failed to create unsigned token: %v",
			err,
		)
	}

	parsedClaims, err := issuer.Parse(tokenString)

	if err == nil {
		t.Fatal("expected parse error")
	}

	if parsedClaims != nil {
		t.Fatal("expected nil claims")
	}
}
