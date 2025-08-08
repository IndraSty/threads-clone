package services

import (
	"fmt"

	"github.com/BangNdraa/threads-clone/backend/auth-service/internal/database"
	"github.com/BangNdraa/threads-clone/backend/auth-service/internal/models"
	"github.com/BangNdraa/threads-clone/backend/auth-service/internal/utils"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo *database.UserRepository
}

func NewUserService(userRepo *database.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return user, nil
}

func (s *UserService) UpdateProfile(userID uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update user profile
	updatedUser, err := s.userRepo.Update(userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}
	if updatedUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Remove password hash from response
	updatedUser.PasswordHash = ""

	return updatedUser, nil
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return user, nil
}
