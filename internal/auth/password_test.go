package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "correct horse battery staple"

	firstHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	secondHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password a second time: %v", err)
	}

	if firstHash == secondHash {
		t.Fatal("expected unique salts to produce different hashes")
	}
	if !strings.HasPrefix(firstHash, "$argon2id$v=19$") {
		t.Fatalf("expected Argon2id encoded hash, got %q", firstHash)
	}

	valid, err := VerifyPassword(password, firstHash)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if !valid {
		t.Fatal("expected password to verify")
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	valid, err := VerifyPassword("wrong password", hash)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if valid {
		t.Fatal("expected wrong password to be rejected")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	valid, err := VerifyPassword("password", "not-a-password-hash")
	if err != ErrInvalidPasswordHash {
		t.Fatalf("expected invalid hash error, got %v", err)
	}
	if valid {
		t.Fatal("expected malformed hash to be rejected")
	}
}
