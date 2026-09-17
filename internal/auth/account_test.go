package auth

import (
	"context"
	"errors"
	"testing"

	repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"
)

type accountUserCreator struct {
	params repo.CreateUserParams
	user   repo.User
}

func (f *accountUserCreator) CreateUser(_ context.Context, params repo.CreateUserParams) (repo.User, error) {
	f.params = params
	return repo.User{ID: 1}, nil
}

func (f *accountUserCreator) GetUserByEmail(context.Context, string) (repo.User, error) {
	return f.user, nil
}

func TestCreateAccountNormalizesAndHashesInput(t *testing.T) {
	creator := &accountUserCreator{}
	service := NewAccountService(creator)

	_, err := service.CreateAccount(context.Background(), AccountInput{
		Name:        "  Ada Lovelace ",
		Email:       " ADA@example.com ",
		Password:    "correct horse battery staple",
		PhoneNumber: " +441234567890 ",
		Country:     "gb",
		Username:    " Ada_1 ",
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	if creator.params.Name != "Ada Lovelace" || creator.params.Email != "ada@example.com" {
		t.Fatalf("input was not normalized: %+v", creator.params)
	}
	if creator.params.PhoneNumber.String != "+441234567890" || creator.params.Country != "GB" || creator.params.Username != "ada_1" {
		t.Fatalf("normalized fields are incorrect: %+v", creator.params)
	}
	if creator.params.PasswordHash == "correct horse battery staple" {
		t.Fatal("plaintext password was persisted")
	}
	valid, err := VerifyPassword("correct horse battery staple", creator.params.PasswordHash)
	if err != nil || !valid {
		t.Fatalf("persisted password hash does not verify: valid=%v err=%v", valid, err)
	}
}

func TestCreateAccountRejectsInvalidInputBeforePersistence(t *testing.T) {
	creator := &accountUserCreator{}
	service := NewAccountService(creator)

	_, err := service.CreateAccount(context.Background(), AccountInput{
		Name:        "Ada Lovelace",
		Email:       "ada@example.com",
		Password:    "short",
		PhoneNumber: "+441234567890",
		Country:     "GB",
		Username:    "ada_1",
	})
	if err == nil {
		t.Fatal("expected invalid password error")
	}
	if creator.params.PasswordHash != "" {
		t.Fatal("expected invalid input not to reach persistence")
	}
}

func TestLoginReturnsTokenForVerifiedUser(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	passwordHash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	creator := &accountUserCreator{user: repo.User{ID: 1, Email: "ada@example.com", PasswordHash: passwordHash, EmailVerified: true}}
	result, err := NewAccountService(creator).Login(context.Background(), LoginInput{
		Email:    " ADA@example.com ",
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.Token == "" || result.User.ID != 1 {
		t.Fatalf("unexpected login result: %+v", result)
	}
}

func TestLoginRejectsUnverifiedUser(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	passwordHash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	creator := &accountUserCreator{user: repo.User{PasswordHash: passwordHash}}
	_, err = NewAccountService(creator).Login(context.Background(), LoginInput{Password: "correct horse battery staple"})
	if !errors.Is(err, ErrEmailNotVerified) {
		t.Fatalf("expected unverified error, got %v", err)
	}
}
