package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory      uint32 = 64 * 1024
	argon2Iterations  uint32 = 3
	argon2Parallelism uint8  = 2
	argon2SaltLength         = 16
	argon2KeyLength          = 32
)

var ErrInvalidPasswordHash = errors.New("invalid password hash")

func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory,
		argon2Iterations,
		argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	memory, iterations, parallelism, salt, expectedKey, err := parsePasswordHash(encodedHash)
	if err != nil {
		return false, err
	}

	actualKey := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedKey)))
	return subtle.ConstantTimeCompare(actualKey, expectedKey) == 1, nil
}

func parsePasswordHash(encodedHash string) (uint32, uint32, uint8, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	parameters := strings.Split(parts[3], ",")
	if len(parameters) != 3 {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	values := make(map[string]string, len(parameters))
	for _, parameter := range parameters {
		key, value, ok := strings.Cut(parameter, "=")
		if !ok || (key != "m" && key != "t" && key != "p") {
			return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
		}
		if _, exists := values[key]; exists {
			return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
		}
		values[key] = value
	}

	memory, err := strconv.ParseUint(values["m"], 10, 32)
	if err != nil || memory == 0 {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}
	iterations, err := strconv.ParseUint(values["t"], 10, 32)
	if err != nil || iterations == 0 {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}
	parallelism, err := strconv.ParseUint(values["p"], 10, 8)
	if err != nil || parallelism == 0 {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) == 0 {
		return 0, 0, 0, nil, nil, ErrInvalidPasswordHash
	}

	return uint32(memory), uint32(iterations), uint8(parallelism), salt, key, nil
}
