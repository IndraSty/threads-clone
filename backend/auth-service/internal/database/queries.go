package database

import (
	"database/sql"
	"fmt"

	"github.com/IndraSty/threads-clone/backend/auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

const (
	createUserQuery = `
		INSERT INTO users (username, display_name, email, password_hash, oauth_providers)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, display_name, email, bio, profile_image_url, oauth_providers, created_at`

	getUserByIDQuery = `
		SELECT id, username, display_name, email, bio, profile_image_url, oauth_providers, created_at
		FROM users WHERE id = $1`

	getUserByEmailQuery = `
		SELECT id, username, display_name, email, password_hash, bio, profile_image_url, oauth_providers, created_at
		FROM users WHERE email = $1`

	getUserByUsernameQuery = `
		SELECT id, username, display_name, email, password_hash, bio, profile_image_url, oauth_providers, created_at
		FROM users WHERE username = $1`

	getUserByOAuthQuery = `
		SELECT id, username, display_name, email, password_hash, bio, profile_image_url, oauth_providers, created_at
		FROM users WHERE oauth_providers->$1->>'id' = $2`

	updateUserQuery = `
		UPDATE users SET 
			display_name = COALESCE($2, display_name),
			bio = COALESCE($3, bio),
			profile_image_url = COALESCE($4, profile_image_url)
		WHERE id = $1
		RETURNING id, username, display_name, email, bio, profile_image_url, oauth_providers, created_at`

	updateUserOAuthQuery = `
		UPDATE users SET oauth_providers = $2
		WHERE id = $1
		RETURNING id, username, display_name, email, bio, profile_image_url, oauth_providers, created_at`

	checkEmailExistsQuery    = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	checkUsernameExistsQuery = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
)

func (r *UserRepository) Create(user *models.User) (*models.User, error) {
	var createdUser models.User

	err := r.db.QueryRowx(
		createUserQuery,
		user.Username,
		user.DisplayName,
		user.Email,
		user.PasswordHash,
		user.OAuthProviders,
	).StructScan(&createdUser)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &createdUser, nil
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowx(getUserByIDQuery, id).StructScan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowx(getUserByEmailQuery, email).StructScan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowx(getUserByUsernameQuery, username).StructScan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByOAuth(provider, providerID string) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowx(getUserByOAuthQuery, provider, providerID).StructScan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by OAuth: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(id uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowx(
		updateUserQuery,
		id,
		req.DisplayName,
		req.Bio,
		req.ProfileImageURL,
	).StructScan(&user)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateOAuth(id uuid.UUID, oauthData *models.OAuthData) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowx(updateUserOAuthQuery, id, oauthData).StructScan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update user OAuth: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(checkEmailExistsQuery, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) UsernameExists(username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(checkUsernameExistsQuery, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}
