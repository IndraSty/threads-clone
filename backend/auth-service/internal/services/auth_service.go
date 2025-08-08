package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/IndraSty/threads-clone/backend/auth-service/internal/database"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/models"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/utils"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo     *database.UserRepository
	jwtManager   *utils.JWTManager
	oauthManager *utils.OAuthManager
}

func NewAuthService(
	userRepo *database.UserRepository,
	jwtManager *utils.JWTManager,
	oauthManager *utils.OAuthManager,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		jwtManager:   jwtManager,
		oauthManager: oauthManager,
	}
}

func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if email already exists
	emailExists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if emailExists {
		return nil, fmt.Errorf("email already registered")
	}

	// Check if username already exists
	usernameExists, err := s.userRepo.UsernameExists(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if usernameExists {
		return nil, fmt.Errorf("username already taken")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	createdUser, err := s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(createdUser.ID, createdUser.Username, createdUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{
		User:        createdUser,
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwtManager.GetExpiresIn(),
	}, nil
}

func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := utils.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &models.AuthResponse{
		User:        user,
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwtManager.GetExpiresIn(),
	}, nil
}

func (s *AuthService) GetGoogleAuthURL(state string) string {
	return s.oauthManager.GetGoogleAuthURL(state)
}

func (s *AuthService) GetFacebookAuthURL(state string) string {
	return s.oauthManager.GetFacebookAuthURL(state)
}

func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*models.AuthResponse, error) {
	// Exchange code for user info
	userInfo, err := s.oauthManager.ExchangeGoogleCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange Google code: %w", err)
	}

	return s.handleOAuthUser(userInfo)
}

func (s *AuthService) HandleFacebookCallback(ctx context.Context, code string) (*models.AuthResponse, error) {
	// Exchange code for user info
	userInfo, err := s.oauthManager.ExchangeFacebookCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange Facebook code: %w", err)
	}

	return s.handleOAuthUser(userInfo)
}

func (s *AuthService) handleOAuthUser(userInfo *models.OAuthUserInfo) (*models.AuthResponse, error) {
	// Check if user exists with this OAuth provider
	existingUser, err := s.userRepo.GetByOAuth(userInfo.Provider, userInfo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check OAuth user: %w", err)
	}

	var user *models.User

	if existingUser != nil {
		// User exists, update their info if needed
		user = existingUser
	} else {
		// Check if user exists with the same email
		emailUser, err := s.userRepo.GetByEmail(userInfo.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email user: %w", err)
		}

		if emailUser != nil {
			// Link OAuth to existing email account
			user, err = s.linkOAuthToUser(emailUser, userInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to link OAuth: %w", err)
			}
		} else {
			// Create new user
			user, err = s.createOAuthUser(userInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to create OAuth user: %w", err)
			}
		}
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &models.AuthResponse{
		User:        user,
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwtManager.GetExpiresIn(),
	}, nil
}

func (s *AuthService) createOAuthUser(userInfo *models.OAuthUserInfo) (*models.User, error) {
	// Generate unique username from name and email
	username := s.generateUniqueUsername(userInfo.Name, userInfo.Email)

	// Create OAuth data
	oauthData := &models.OAuthData{}
	oauthUserData := &models.OAuthUserData{
		ID:    userInfo.ID,
		Email: userInfo.Email,
	}

	switch userInfo.Provider {
	case "google":
		oauthData.Google = oauthUserData
	case "facebook":
		oauthData.Facebook = oauthUserData
	}

	// Create user
	user := &models.User{
		Username:       username,
		DisplayName:    userInfo.Name,
		Email:          userInfo.Email,
		PasswordHash:   "", // OAuth users don't have password
		OAuthProviders: oauthData,
	}

	if userInfo.Picture != "" {
		user.ProfileImageURL = &userInfo.Picture
	}

	createdUser, err := s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth user: %w", err)
	}

	return createdUser, nil
}

func (s *AuthService) linkOAuthToUser(user *models.User, userInfo *models.OAuthUserInfo) (*models.User, error) {
	// Initialize OAuth data if nil
	if user.OAuthProviders == nil {
		user.OAuthProviders = &models.OAuthData{}
	}

	oauthUserData := &models.OAuthUserData{
		ID:    userInfo.ID,
		Email: userInfo.Email,
	}

	// Add new OAuth provider
	switch userInfo.Provider {
	case "google":
		user.OAuthProviders.Google = oauthUserData
	case "facebook":
		user.OAuthProviders.Facebook = oauthUserData
	}

	// Update user with OAuth data
	updatedUser, err := s.userRepo.UpdateOAuth(user.ID, user.OAuthProviders)
	if err != nil {
		return nil, fmt.Errorf("failed to update user OAuth: %w", err)
	}

	return updatedUser, nil
}

func (s *AuthService) generateUniqueUsername(name, email string) string {
	// Clean name and create base username
	baseUsername := strings.ToLower(strings.ReplaceAll(name, " ", ""))
	if baseUsername == "" {
		// Use email prefix if name is empty
		emailParts := strings.Split(email, "@")
		baseUsername = emailParts[0]
	}

	// Remove special characters
	baseUsername = strings.ReplaceAll(baseUsername, ".", "")
	baseUsername = strings.ReplaceAll(baseUsername, "-", "")

	// Try base username first
	username := baseUsername
	counter := 1

	// Keep trying until we find a unique username
	for {
		exists, err := s.userRepo.UsernameExists(username)
		if err == nil && !exists {
			break
		}

		// Add counter to make it unique
		username = fmt.Sprintf("%s%d", baseUsername, counter)
		counter++

		// Prevent infinite loop
		if counter > 1000 {
			username = fmt.Sprintf("%s%s", baseUsername, uuid.New().String()[:8])
			break
		}
	}

	return username
}

func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	// Validate JWT token
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get user from database
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return user, nil
}
