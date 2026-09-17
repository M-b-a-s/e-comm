package users

import (
	"context"
	"fmt"

	repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"
)

type repository interface {
	GetUserByEmail(context.Context, string) (repo.User, error)
	GetUserByID(context.Context, int64) (repo.User, error)
	ListUsers(context.Context, repo.ListUsersParams) ([]repo.User, error)
	VerifyUserEmail(context.Context, int64) error
}

type Service struct {
	repository repository
}

func (s *Service) GetByEmail(ctx context.Context, email string) (repo.User, error) {
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		return repo.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetByID(ctx context.Context, id int64) (repo.User, error) {
	user, err := s.repository.GetUserByID(ctx, id)
	if err != nil {
		return repo.User{}, fmt.Errorf("get user %d: %w", id, err)
	}
	return user, nil
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]repo.User, error) {
	if limit < 0 || offset < 0 {
		return nil, fmt.Errorf("list users: limit and offset cannot be negative")
	}

	users, err := s.repository.ListUsers(ctx, repo.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *Service) VerifyEmail(ctx context.Context, id int64) error {
	if err := s.repository.VerifyUserEmail(ctx, id); err != nil {
		return fmt.Errorf("verify email for user %d: %w", id, err)
	}
	return nil
}
