package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/IndraSty/threads-clone/backend/auth-service/configs"
	"github.com/IndraSty/threads-clone/backend/auth-service/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

type OAuthManager struct {
	googleConfig   *oauth2.Config
	facebookConfig *oauth2.Config
}

func NewOAuthManager(cfg *configs.Config) *OAuthManager {
	return &OAuthManager{
		googleConfig: &oauth2.Config{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			RedirectURL:  cfg.OAuth.Google.RedirectURL,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		},
		facebookConfig: &oauth2.Config{
			ClientID:     cfg.OAuth.Facebook.ClientID,
			ClientSecret: cfg.OAuth.Facebook.ClientSecret,
			RedirectURL:  cfg.OAuth.Facebook.RedirectURL,
			Scopes:       []string{"email", "public_profile"},
			Endpoint:     facebook.Endpoint,
		},
	}
}

func (o *OAuthManager) GetGoogleAuthURL(state string) string {
	return o.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (o *OAuthManager) GetFacebookAuthURL(state string) string {
	return o.facebookConfig.AuthCodeURL(state)
}

func (o *OAuthManager) ExchangeGoogleCode(ctx context.Context, code string) (*models.OAuthUserInfo, error) {
	token, err := o.googleConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange Google code: %w", err)
	}

	userInfo, err := o.getGoogleUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get Google user info: %w", err)
	}

	userInfo.Provider = "google"
	return userInfo, nil
}

func (o *OAuthManager) ExchangeFacebookCode(ctx context.Context, code string) (*models.OAuthUserInfo, error) {
	token, err := o.facebookConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange Facebook code: %w", err)
	}

	userInfo, err := o.getFacebookUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get Facebook user info: %w", err)
	}

	userInfo.Provider = "facebook"
	return userInfo, nil
}

func (o *OAuthManager) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*models.OAuthUserInfo, error) {
	client := o.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo models.OAuthUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (o *OAuthManager) getFacebookUserInfo(ctx context.Context, token *oauth2.Token) (*models.OAuthUserInfo, error) {
	client := o.facebookConfig.Client(ctx, token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email,picture")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var fbUser struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.Unmarshal(body, &fbUser); err != nil {
		return nil, err
	}

	userInfo := &models.OAuthUserInfo{
		ID:      fbUser.ID,
		Name:    fbUser.Name,
		Email:   fbUser.Email,
		Picture: fbUser.Picture.Data.URL,
	}

	return userInfo, nil
}
