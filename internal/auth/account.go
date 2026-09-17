package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const minimumPasswordLength = 12

var (
	ErrInvalidAccountInput = errors.New("invalid account input")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailNotVerified    = errors.New("email is not verified")
	ErrJWTSecretMissing    = errors.New("JWT_SECRET is not configured")
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

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	User  repo.User
	Token string
}

type AdminInput struct {
	Name        string
	Email       string
	Password    string
	PhoneNumber string
	Country     string
	Username    string
}

type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

type userCreator interface {
	CreateUser(context.Context, repo.CreateUserParams) (repo.User, error)
	CreateAdminUser(context.Context, repo.CreateAdminUserParams) (repo.User, error)
	GetUserByEmail(context.Context, string) (repo.User, error)
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

func (s *AccountService) BootstrapAdmin(ctx context.Context, input AdminInput) (repo.User, error) {
	accountInput, err := normalizeAccountInput(AccountInput{
		Name:        input.Name,
		Email:       input.Email,
		Password:    input.Password,
		PhoneNumber: input.PhoneNumber,
		Country:     input.Country,
		Username:    input.Username,
	})
	if err != nil {
		return repo.User{}, err
	}

	passwordHash, err := HashPassword(accountInput.Password)
	if err != nil {
		return repo.User{}, fmt.Errorf("hash admin password: %w", err)
	}

	user, err := s.users.CreateAdminUser(ctx, repo.CreateAdminUserParams{
		Name:         accountInput.Name,
		Email:        accountInput.Email,
		PasswordHash: passwordHash,
		PhoneNumber:  pgtype.Text{String: accountInput.PhoneNumber, Valid: true},
		Country:      accountInput.Country,
		Username:     accountInput.Username,
	})
	if err != nil {
		return repo.User{}, fmt.Errorf("bootstrap admin: %w", err)
	}
	return user, nil
}

func (s *AccountService) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	valid, err := VerifyPassword(input.Password, user.PasswordHash)
	if err != nil || !valid {
		return LoginResult{}, ErrInvalidCredentials
	}
	if !user.EmailVerified {
		return LoginResult{}, ErrEmailNotVerified
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return LoginResult{}, ErrJWTSecretMissing
	}

	now := time.Now()
	claims := Claims{
		Role:      string(user.Role),
		Subject:   fmt.Sprintf("%d", user.ID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return LoginResult{}, fmt.Errorf("sign login token: %w", err)
	}

	return LoginResult{User: user, Token: token}, nil
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
