package otp

import (
	"errors"
	"regexp"
	"testing"
	"time"
)

type memoryStore struct {
	hash     string
	expires  time.Time
	attempts int
	deleted  bool
}

func (s *memoryStore) Save(_ string, hashedCode string, expiresAt time.Time) error {
	s.hash = hashedCode
	s.expires = expiresAt
	return nil
}

func (s *memoryStore) Get(_ string) (string, time.Time, int, error) {
	if s.deleted {
		return "", time.Time{}, 0, errors.New("not found")
	}
	return s.hash, s.expires, s.attempts, nil
}

func (s *memoryStore) IncrementAttempts(_ string) error {
	s.attempts++
	return nil
}

func (s *memoryStore) Delete(_ string) error {
	s.deleted = true
	return nil
}

type recordingSender struct {
	toEmail string
	code    string
}

func (s *recordingSender) SendOTP(toEmail, code string) error {
	s.toEmail = toEmail
	s.code = code
	return nil
}

func TestGenerateReturnsSixDigitCode(t *testing.T) {
	code, err := Generate()
	if err != nil {
		t.Fatalf("generate OTP: %v", err)
	}
	if !regexp.MustCompile(`^[0-9]{6}$`).MatchString(code) {
		t.Fatalf("expected six-digit code, got %q", code)
	}
}

func TestIssueAndVerify(t *testing.T) {
	t.Setenv("OTP_HMAC_SECRET", "test-secret")
	store := &memoryStore{}
	sender := &recordingSender{}

	if err := IssueAndSend(store, sender, "ada@example.com"); err != nil {
		t.Fatalf("issue OTP: %v", err)
	}
	if sender.toEmail != "ada@example.com" || sender.code == "" {
		t.Fatalf("sender did not receive OTP: %+v", sender)
	}
	if store.hash == sender.code {
		t.Fatal("stored OTP must be hashed")
	}

	valid, err := Verify(store, "ada@example.com", sender.code)
	if err != nil || !valid {
		t.Fatalf("verify OTP: valid=%v err=%v", valid, err)
	}
	if !store.deleted {
		t.Fatal("verified OTP should be deleted")
	}
}

func TestVerifyIncrementsAttemptsForInvalidCode(t *testing.T) {
	t.Setenv("OTP_HMAC_SECRET", "test-secret")
	store := &memoryStore{hash: "different", expires: time.Now().Add(time.Minute)}

	valid, err := Verify(store, "ada@example.com", "123456")
	if err != nil || valid {
		t.Fatalf("expected invalid OTP: valid=%v err=%v", valid, err)
	}
	if store.attempts != 1 {
		t.Fatalf("expected one failed attempt, got %d", store.attempts)
	}
}
