package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
	"yp-diploma/internal/app/config"
	"yp-diploma/internal/app/model"
	"yp-diploma/internal/app/util"

	"github.com/google/uuid"
)

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
		return "", fmt.Errorf("error creating session key for user: %s", user.Name)
	}
	return cryptKey, nil
}

func (s *Service) LoginUser(ctx context.Context, name, password string) (string, error) {
	user, err := s.repo.GetUserID(ctx, name)
	if err != nil {
		return "", err
	}

	if !CheckPasswd(password, user.HashedPasswd) {
		return "", config.ErrUserInvalidPassword
	}

	cryptKey := s.genSessKey(ctx, user)
	if cryptKey == "" {
		return "", fmt.Errorf("error creating session key for user: %s", user.Name)
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
		UserID:  user.ID,
		Expires: time.Now().Add(config.SessionKeyDuration),
	}
	err = s.repo.AddSessKey(ctx, key)
	if err != nil {
		return ""
	}
	cryptKey, _ := util.EncodeString(util.SignString(sessKey))
	return cryptKey
}

func CheckPasswd(password, hash string) bool {
	pwdHash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(pwdHash[:]) == hash
}
