package otp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"
)

const otpTTL = 10 * time.Minute

var ErrSecretNotConfigured = errors.New("OTP_HMAC_SECRET is not configured")

type Store interface {
	Save(email, hashedCode string, expiresAt time.Time) error
	Get(email string) (hashedCode string, expiresAt time.Time, attempts int, err error)
	IncrementAttempts(email string) error
	Delete(email string) error
}

type Sender interface {
	SendOTP(toEmail, code string) error
}

func Generate() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func IssueAndSend(store Store, sender Sender, toEmail string) error {
	code, err := Generate()
	if err != nil {
		return err
	}

	codeHash, err := hash(code)
	if err != nil {
		return err
	}
	if err := store.Save(toEmail, codeHash, time.Now().Add(otpTTL)); err != nil {
		return err
	}

	return sender.SendOTP(toEmail, code)
}

func Verify(store Store, email, submittedCode string) (bool, error) {
	hashedCode, expiresAt, attempts, err := store.Get(email)
	if err != nil {
		return false, err
	}

	if attempts >= 5 {
		return false, fmt.Errorf("too many attempts")
	}

	if time.Now().After(expiresAt) {
		if err := store.Delete(email); err != nil {
			return false, fmt.Errorf("delete expired OTP: %w", err)
		}
		return false, fmt.Errorf("code expired")
	}

	submittedHash, err := hash(submittedCode)
	if err != nil {
		return false, err
	}
	if subtle.ConstantTimeCompare([]byte(submittedHash), []byte(hashedCode)) != 1 {
		if err := store.IncrementAttempts(email); err != nil {
			return false, fmt.Errorf("increment OTP attempts: %w", err)
		}
		return false, nil
	}

	if err := store.Delete(email); err != nil {
		return false, fmt.Errorf("delete OTP: %w", err)
	}
	return true, nil
}

func hash(code string) (string, error) {
	secret := os.Getenv("OTP_HMAC_SECRET")
	if secret == "" {
		return "", ErrSecretNotConfigured
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
