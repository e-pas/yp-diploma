package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/model"
	"yp-diploma/internal/app/repository"
	"yp-diploma/internal/app/util"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) RegisterUser(ctx context.Context, name, password string) (string, error) {
	pwdHash := sha256.Sum256([]byte(password))
	user := model.User{
		ID:           uuid.New().String(),
		Name:         name,
		HashedPasswd: hex.EncodeToString(pwdHash[:]),
	}
	err := s.repo.AddUser(ctx, user)
	if err != nil {
		return "", err
	}

	cryptKey := s.genSessKey(ctx, user)
	if cryptKey == "" {
		return "", errors.New(fmt.Sprintf("error creating session key for user: %s", user.Name))
	}
	return cryptKey, nil
}

func (s *Service) LoginUser(ctx context.Context, name, password string) (string, error) {
	user, err := s.repo.GetUserID(ctx, name)
	if err != nil {
		return "", err
	}

	if !CheckPasswd(password, user.HashedPasswd) {
		return "", errors.New("invalid password")
	}

	cryptKey := s.genSessKey(ctx, user)
	if cryptKey == "" {
		return "", errors.New(fmt.Sprintf("error creating session key for user: %s", user.Name))
	}
	return cryptKey, nil
}

func (s *Service) genSessKey(ctx context.Context, user model.User) string {
	sessKey, err := util.GetRandHexString(16)
	if err != nil {
		return ""
	}
	key := model.SessKey{
		ID:      sessKey,
		User_ID: user.ID,
		Expires: time.Now().Add(config.SessionKeyDuration),
	}
	err = s.repo.AddSessKey(ctx, key)
	if err != nil {
		return ""
	}
	cryptKey, err := util.EncodeString(util.SignString(sessKey))
	return cryptKey
}

func CheckPasswd(password, hash string) bool {
	pwdHash := sha256.Sum256([]byte(password))
	return hash == hex.EncodeToString(pwdHash[:])
}
