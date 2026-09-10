package auth

import (
	"context"
	"testing"

	repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"
)

type accountUserCreator struct {
	params repo.CreateUserParams
}

func (f *accountUserCreator) CreateUser(_ context.Context, params repo.CreateUserParams) (repo.User, error) {
	f.params = params
	return repo.User{ID: 1}, nil
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
