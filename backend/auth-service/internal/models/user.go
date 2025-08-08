package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Username        string     `json:"username" db:"username" validate:"required,min=3,max=50"`
	DisplayName     string     `json:"display_name" db:"display_name" validate:"required,min=1,max=100"`
	Email           string     `json:"email" db:"email" validate:"required,email"`
	PasswordHash    string     `json:"-" db:"password_hash"`
	Bio             *string    `json:"bio" db:"bio"`
	ProfileImageURL *string    `json:"profile_image_url" db:"profile_image_url"`
	OAuthProviders  *OAuthData `json:"oauth_providers,omitempty" db:"oauth_providers"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type OAuthData struct {
	Google   *OAuthUserData `json:"google,omitempty"`
	Facebook *OAuthUserData `json:"facebook,omitempty"`
}

type OAuthUserData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Implement driver.Valuer interface for OAuthData
func (o OAuthData) Value() (driver.Value, error) {
	return json.Marshal(o)
}

// Implement sql.Scanner interface for OAuthData
func (o *OAuthData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, o)
}

// Request/Response models
type RegisterRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	DisplayName     *string `json:"display_name,omitempty" validate:"omitempty,min=1,max=100"`
	Bio             *string `json:"bio,omitempty"`
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
}

type AuthResponse struct {
	User        *User  `json:"user"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type OAuthUserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Provider string `json:"provider"`
}
