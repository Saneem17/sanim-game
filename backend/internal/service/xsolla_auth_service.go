package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const xsollaUsersInfoURL = "https://login.xsolla.com/api/users/me"

// XsollaClaims — maklumat user yang diambil selepas token disahkan
type XsollaClaims struct {
	Subject              string // User ID (UUID dari Xsolla)
	XsollaLoginProjectID string // Project ID untuk pastikan token milik project kita
}

// XsollaAuthService — verify token dengan panggil Xsolla API terus (tanpa JWKS)
type XsollaAuthService struct {
	ExpectedProjID string
}

// NewXsollaAuthService — buat XsollaAuthService dengan project ID yang dijangka
func NewXsollaAuthService(projectID string) *XsollaAuthService {
	return &XsollaAuthService{
		ExpectedProjID: projectID,
	}
}

// VerifyToken — sahkan token dengan hantar ke Xsolla /users/me
// Kalau Xsolla balas 200, token valid. Kalau 401, token expired atau tak sah.
func (s *XsollaAuthService) VerifyToken(ctx context.Context, tokenString string) (*XsollaClaims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, xsollaUsersInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("xsolla auth check failed with status %d", resp.StatusCode)
	}

	var userInfo struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	if userInfo.ID == "" {
		return nil, errors.New("empty user id in xsolla response")
	}

	// Decode payload JWT (tanpa verify signature) untuk dapat project ID
	projID, err := extractProjectID(tokenString)
	if err != nil {
		return nil, err
	}

	if projID != s.ExpectedProjID {
		return nil, errors.New("invalid xsolla_login_project_id")
	}

	return &XsollaClaims{
		Subject:              userInfo.ID,
		XsollaLoginProjectID: projID,
	}, nil
}

// extractProjectID — decode bahagian payload JWT (base64) untuk ambil xsolla_login_project_id
// Kita tak verify signature di sini — token dah disahkan oleh Xsolla API di atas
func extractProjectID(tokenString string) (string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) < 2 {
		return "", errors.New("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	var claims struct {
		ProjectID string `json:"xsolla_login_project_id"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", err
	}

	return claims.ProjectID, nil
}
