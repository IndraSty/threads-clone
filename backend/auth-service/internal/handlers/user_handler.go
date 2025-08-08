package handlers

import (
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/middleware"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/models"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/services"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile returns current user's profile
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse(
			"UNAUTHORIZED",
			"User not authenticated",
			err.Error(),
		))
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse(
				"USER_NOT_FOUND",
				"User not found",
				nil,
			))
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(
			"GET_PROFILE_FAILED",
			"Failed to get user profile",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse(
		"Profile retrieved successfully",
		user,
	))
}

// UpdateProfile updates current user's profile
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse(
			"UNAUTHORIZED",
			"User not authenticated",
			err.Error(),
		))
	}

	var req models.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
	}

	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		// Check for validation errors
		if validationErrors := utils.FormatValidationErrors(err); len(validationErrors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
				"VALIDATION_ERROR",
				"Validation failed",
				validationErrors,
			))
		}

		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse(
				"USER_NOT_FOUND",
				"User not found",
				nil,
			))
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(
			"UPDATE_PROFILE_FAILED",
			"Failed to update user profile",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse(
		"Profile updated successfully",
		user,
	))
}

// GetUserByUsername returns user profile by username
func (h *UserHandler) GetUserByUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse(
			"INVALID_REQUEST",
			"Username parameter is required",
			nil,
		))
	}

	user, err := h.userService.GetUserByUsername(username)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse(
				"USER_NOT_FOUND",
				"User not found",
				nil,
			))
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse(
			"GET_USER_FAILED",
			"Failed to get user",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse(
		"User retrieved successfully",
		user,
	))
}
