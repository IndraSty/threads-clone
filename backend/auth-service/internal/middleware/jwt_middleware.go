package middleware

import (
	"strings"

	"github.com/BangNdraa/threads-clone/backend/auth-service/internal/models"
	"github.com/BangNdraa/threads-clone/backend/auth-service/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// JWTMiddleware validates JWT tokens and adds user info to context
func (m *AuthMiddleware) JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse(
				"UNAUTHORIZED",
				"Missing authorization header",
				nil,
			))
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse(
				"UNAUTHORIZED",
				"Invalid authorization header format",
				nil,
			))
		}

		token := tokenParts[1]

		// Validate token and get user
		user, err := m.authService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse(
				"UNAUTHORIZED",
				"Invalid or expired token",
				err.Error(),
			))
		}

		// Add user to context
		c.Locals("user", user)
		c.Locals("userID", user.ID)
		c.Locals("username", user.Username)

		return c.Next()
	}
}

// OptionalJWTMiddleware validates JWT tokens if present but doesn't require them
func (m *AuthMiddleware) OptionalJWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next() // Continue without user context
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			return c.Next() // Continue without user context
		}

		token := tokenParts[1]

		// Validate token and get user
		user, err := m.authService.ValidateToken(token)
		if err != nil {
			return c.Next() // Continue without user context
		}

		// Add user to context
		c.Locals("user", user)
		c.Locals("userID", user.ID)
		c.Locals("username", user.Username)

		return c.Next()
	}
}

// GetUserFromContext extracts user from fiber context
func GetUserFromContext(c *fiber.Ctx) (*models.User, error) {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "User not found in context")
	}
	return user, nil
}

// GetUserIDFromContext extracts user ID from fiber context
func GetUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User ID not found in context")
	}
	return userID, nil
}

// GetUsernameFromContext extracts username from fiber context
func GetUsernameFromContext(c *fiber.Ctx) (string, error) {
	username, ok := c.Locals("username").(string)
	if !ok {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Username not found in context")
	}
	return username, nil
}
