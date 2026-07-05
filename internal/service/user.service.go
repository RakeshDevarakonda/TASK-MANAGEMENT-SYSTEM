package service

import (
	"strings"
	"task-system/internal/models"
	"task-system/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
}

func NewUserService(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository) *UserService {
	return &UserService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// UpdateProfile updates the profile info for a user
func (s *UserService) UpdateProfile(userID string, req models.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		user.Name = strings.TrimSpace(*req.Name)
	}

	if req.Email != nil {
		user.Email = strings.ToLower(strings.TrimSpace(*req.Email))
	}

	passwordChanged := false
	if req.Password != nil {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(passwordHash)
		passwordChanged = true
	}

	err = s.userRepo.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	if passwordChanged {
		_ = s.sessionRepo.DeleteByUserID(userID)
	}

	return user, nil
}

// DeleteProfile deletes a user and cleans up all active session records
func (s *UserService) DeleteProfile(userID string) error {
	// 1. Invalidate sessions
	_ = s.sessionRepo.DeleteByUserID(userID)

	// 2. Delete user
	return s.userRepo.DeleteUser(userID)
}
