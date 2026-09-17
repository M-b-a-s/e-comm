package postgresql

import (
	"context"
	"time"

	repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"
	"github/M-b-a-s/e-comm/internal/otp"

	"github.com/jackc/pgx/v5/pgtype"
)

var _ otp.Store = (*OTPStore)(nil)

type OTPStore struct {
	queries *repo.Queries
}

func NewOTPStore(queries *repo.Queries) *OTPStore {
	return &OTPStore{queries: queries}
}

func (s *OTPStore) Save(email, hashedCode string, expiresAt time.Time) error {
	return s.queries.SaveOTP(context.Background(), repo.SaveOTPParams{
		Email:      email,
		HashedCode: hashedCode,
		ExpiresAt:  pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

func (s *OTPStore) Get(email string) (string, time.Time, int, error) {
	otp, err := s.queries.GetOTP(context.Background(), email)
	if err != nil {
		return "", time.Time{}, 0, err
	}
	return otp.HashedCode, otp.ExpiresAt.Time, int(otp.Attempts), nil
}

func (s *OTPStore) IncrementAttempts(email string) error {
	return s.queries.IncrementOTPAttempts(context.Background(), email)
}

func (s *OTPStore) Delete(email string) error {
	return s.queries.DeleteOTP(context.Background(), email)
}
