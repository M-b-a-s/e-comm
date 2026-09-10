package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

const minimumPasswordLength = 12

var (
	ErrInvalidAccountInput = errors.New("invalid account input")
	usernamePattern        = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)
)

type AccountInput struct {
	Name        string
	Email       string
	Password    string
	PhoneNumber string
	Country     string
	Username    string
}

type userCreator interface {
	CreateUser(context.Context, repo.CreateUserParams) (repo.User, error)
}

type AccountService struct {
	users userCreator
}

func NewAccountService(users userCreator) *AccountService {
	return &AccountService{users: users}
}

func (s *AccountService) CreateAccount(ctx context.Context, input AccountInput) (repo.User, error) {
	input, err := normalizeAccountInput(input)
	if err != nil {
		return repo.User{}, err
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return repo.User{}, fmt.Errorf("hash account password: %w", err)
	}

	user, err := s.users.CreateUser(ctx, repo.CreateUserParams{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: passwordHash,
		PhoneNumber:  pgtype.Text{String: input.PhoneNumber, Valid: true},
		Country:      input.Country,
		Username:     input.Username,
	})
	if err != nil {
		return repo.User{}, fmt.Errorf("create account: %w", err)
	}

	return user, nil
}

func normalizeAccountInput(input AccountInput) (AccountInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)
	input.Country = strings.ToUpper(strings.TrimSpace(input.Country))
	input.Username = strings.ToLower(strings.TrimSpace(input.Username))

	if input.Name == "" || input.PhoneNumber == "" || input.Password == "" {
		return AccountInput{}, fmt.Errorf("%w: name, phone number, and password are required", ErrInvalidAccountInput)
	}
	if len(input.Password) < minimumPasswordLength {
		return AccountInput{}, fmt.Errorf("%w: password must be at least %d characters", ErrInvalidAccountInput, minimumPasswordLength)
	}
	parsedEmail, err := mail.ParseAddress(input.Email)
	if err != nil || parsedEmail.Address != input.Email {
		return AccountInput{}, fmt.Errorf("%w: email is invalid", ErrInvalidAccountInput)
	}
	if len(input.Country) != 2 || input.Country[0] < 'A' || input.Country[0] > 'Z' || input.Country[1] < 'A' || input.Country[1] > 'Z' {
		return AccountInput{}, fmt.Errorf("%w: country must be a two-letter ISO code", ErrInvalidAccountInput)
	}
	if !usernamePattern.MatchString(input.Username) {
		return AccountInput{}, fmt.Errorf("%w: username must contain 3-30 lowercase letters, numbers, or underscores", ErrInvalidAccountInput)
	}

	return input, nil
}
