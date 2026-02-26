package service

import (
	"context"
	"fmt"
	"my_mdb/internal/errors"

	"github.com/sirupsen/logrus"
)

type AuthService struct {
	users UsersRepo
	log   *logrus.Logger
}

func NewAuthService(users UsersRepo, log *logrus.Logger) *AuthService {
	return &AuthService{users: users, log: log}
}

func (s *AuthService) ValidateUserID(ctx context.Context, userID int) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user_id must be positive", errors.ErrBadRequest)
	}

	ok, err := s.users.Exists(ctx, userID)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("users exists check failed")
		return err
	}
	if !ok {
		return errors.ErrNotFound
	}
	return nil
}

func (s *AuthService) CreateUser(ctx context.Context) (int, error) {
	id, err := s.users.Create(ctx)
	if err != nil {
		s.log.WithError(err).Error("create user failed")
		return 0, err
	}
	return id, nil
}
