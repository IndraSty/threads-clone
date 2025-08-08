package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/IndraSty/threads-clone/backend/auth-service/internal/models"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/services"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *services.AuthService
	oauthStates map[string]time.Time // In production, use Redis or similar
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		oauthStates: make(map[string]time.Time),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		// Check for validation errors
		if validationErrors := utils.FormatValidationErrors(err); len(validationErrors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
				"VALIDATION_ERROR",
				"Validation failed",
				validationErrors,
			))
		}

		// Check for duplicate errors
		errMsg := err.Error()
		if errMsg == "email already registered" {
			return c.Status(fiber.StatusConflict).JSON(models.ErrorResponse(
				"EMAIL_EXISTS",
				"Email already registered",
				nil,
			))
		}
		if errMsg == "username already taken" {
			return c.Status(fiber.StatusConflict).JSON(models.ErrorResponse(
				"USERNAME_EXISTS",
				"Username already taken",
				nil,
			))
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(
			"REGISTRATION_FAILED",
			"Failed to register user",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse(
		"User registered successfully",
		response,
	))
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		// Check for validation errors
		if validationErrors := utils.FormatValidationErrors(err); len(validationErrors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
				"VALIDATION_ERROR",
				"Validation failed",
				validationErrors,
			))
		}

		// Check for invalid credentials
		if err.Error() == "invalid email or password" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse(
				"INVALID_CREDENTIALS",
				"Invalid email or password",
				nil,
			))
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(
			"LOGIN_FAILED",
			"Failed to login",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse(
		"Login successful",
		response,
	))
}

// GoogleLogin initiates Google OAuth flow
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	state := h.generateState()
	authURL := h.authService.GetGoogleAuthURL(state)

	// Store state with expiration (5 minutes)
	h.oauthStates[state] = time.Now().Add(5 * time.Minute)

	return c.JSON(models.SuccessResponse(
		"Google OAuth URL generated",
		map[string]string{
			"auth_url": authURL,
			"state":    state,
		},
	))
}

// GoogleCallback handles Google OAuth callback
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
			"INVALID_CALLBACK",
			"Missing code or state parameter",
			nil,
		))
	}

	// Validate state
	if !h.validateState(state) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
			"INVALID_STATE",
			"Invalid or expired state parameter",
			nil,
		))
	}

	response, err := h.authService.HandleGoogleCallback(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(
			"OAUTH_FAILED",
			"Google OAuth authentication failed",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse(
		"Google OAuth login successful",
		response,
	))
}

// FacebookLogin initiates Facebook OAuth flow
func (h *AuthHandler) FacebookLogin(c *fiber.Ctx) error {
	state := h.generateState()
	authURL := h.authService.GetFacebookAuthURL(state)

	// Store state with expiration (5 minutes)
	h.oauthStates[state] = time.Now().Add(5 * time.Minute)

	return c.JSON(models.SuccessResponse(
		"Facebook OAuth URL generated",
		map[string]string{
			"auth_url": authURL,
			"state":    state,
		},
	))
}

// FacebookCallback handles Facebook OAuth callback
func (h *AuthHandler) FacebookCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
			"INVALID_CALLBACK",
			"Missing code or state parameter",
			nil,
		))
	}

	// Validate state
	if !h.validateState(state) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
			"INVALID_STATE",
			"Invalid or expired state parameter",
			nil,
		))
	}

	response, err := h.authService.HandleFacebookCallback(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(
			"OAUTH_FAILED",
			"Facebook OAuth authentication failed",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse(
		"Facebook OAuth login successful",
		response,
	))
}

// ValidateToken validates JWT token (for other services)
func (h *AuthHandler) ValidateToken(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
			"MISSING_TOKEN",
			"Authorization header is required",
			nil,
		))
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	user, err := h.authService.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse(
			"INVALID_TOKEN",
			"Invalid or expired token",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse(
		"Token is valid",
		user,
	))
}

// generateState generates a random state for OAuth
func (h *AuthHandler) generateState() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// validateState validates OAuth state and cleans up expired states
func (h *AuthHandler) validateState(state string) bool {
	expiry, exists := h.oauthStates[state]
	if !exists {
		return false
	}

	// Clean up expired states
	now := time.Now()
	for s, exp := range h.oauthStates {
		if now.After(exp) {
			delete(h.oauthStates, s)
		}
	}

	if now.After(expiry) {
		delete(h.oauthStates, state)
		return false
	}

	// Remove used state
	delete(h.oauthStates, state)
	return true
}
