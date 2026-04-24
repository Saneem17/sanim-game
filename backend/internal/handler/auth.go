package handler

// auth.go — Handle proses login melalui Xsolla OAuth.
// Flow dia macam ni:
//   1. User klik "Login" kat frontend
//   2. Frontend redirect ke Xsolla login page
//   3. Lepas login, Xsolla bagi "code" balik ke frontend
//   4. Frontend hantar code tu ke sini (/auth/xsolla/callback)
//   5. Kita tukar code jadi access token + ambil info user

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// AuthHandler simpan credentials Xsolla yang diperlukan untuk OAuth token exchange
type AuthHandler struct {
	xsollaClientID string // Client ID dari Xsolla dashboard
	redirectURI    string // URL yang kita daftarkan dalam Xsolla (mesti sama persis)
}

// NewAuthHandler — buat AuthHandler baru dengan client ID dan redirect URI
func NewAuthHandler(xsollaClientID, redirectURI string) *AuthHandler {
	return &AuthHandler{
		xsollaClientID: xsollaClientID,
		redirectURI:    redirectURI,
	}
}

// XsollaCallbackRequest — apa yang frontend hantar: just the authorization code
type XsollaCallbackRequest struct {
	Code string `json:"code"` // Code yang Xsolla bagi lepas user login
}

// XsollaTokenResponse — apa yang Xsolla bagi balik bila kita tukar code jadi token
type XsollaTokenResponse struct {
	AccessToken  string `json:"access_token"`  // Token untuk akses API (jangka pendek)
	RefreshToken string `json:"refresh_token"` // Token untuk renew access token (jangka panjang)
	TokenType    string `json:"token_type"`    // Biasanya "Bearer"
	ExpiresIn    int    `json:"expires_in"`    // Berapa saat token tu valid
}

// XsollaUserClaims — maklumat user yang tersimpan dalam JWT token (selepas decode)
type XsollaUserClaims struct {
	Email    string `json:"email"`    // Email user
	Provider string `json:"provider"` // Login provider (google, steam, etc.)
	Sub      string `json:"sub"`      // Unique user ID dari Xsolla
	Picture  string `json:"picture"`  // URL gambar profil user
	Name     string `json:"name"`     // Nama user
}

// XsollaCallbackResponse — apa yang kita hantar balik ke frontend selepas login berjaya
type XsollaCallbackResponse struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Provider     string `json:"provider"`
	UserID       string `json:"user_id"` // ID unik user (dari field "sub" dalam JWT)
	Picture      string `json:"picture"`
}

// HandleXsollaCallback — main function yang handle POST /auth/xsolla/callback
func (h *AuthHandler) HandleXsollaCallback(w http.ResponseWriter, r *http.Request) {
	// Hanya accept POST request
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode JSON body untuk ambil authorization code
	var req XsollaCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Code tu wajib ada — kalau takde, tak boleh proceed
	if req.Code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// Prepare request untuk hantar ke Xsolla token endpoint
	// Ini standard OAuth2 "authorization code" flow
	form := url.Values{}
	form.Set("grant_type", "authorization_code") // Jenis OAuth flow
	form.Set("client_id", h.xsollaClientID)
	form.Set("redirect_uri", h.redirectURI)
	form.Set("code", req.Code) // Code yang dapat dari frontend

	// Hantar request ke Xsolla untuk tukar code jadi token
	resp, err := http.Post(
		"https://login.xsolla.com/api/oauth2/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		http.Error(w, "failed to exchange code", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Kalau Xsolla bagi error, kita hantar error tu ke frontend
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "xsolla token exchange failed: "+string(body), http.StatusBadGateway)
		return
	}

	// Parse JSON response dari Xsolla untuk ambil token
	var tokenResp XsollaTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		http.Error(w, "failed to parse xsolla token response", http.StatusInternalServerError)
		return
	}

	// Decode JWT token untuk ambil maklumat user (email, nama, etc.)
	// JWT ada 3 bahagian: header.payload.signature — kita ambil payload (bahagian tengah)
	claims, err := decodeJWTClaims(tokenResp.AccessToken)
	if err != nil {
		http.Error(w, "failed to decode user token", http.StatusInternalServerError)
		return
	}

	// Kalau nama takde dalam token, guna email sebagai nama. Kalau email pun takde, guna default
	name := claims.Name
	if name == "" {
		if claims.Email != "" {
			name = claims.Email
		} else {
			name = "Garden Player"
		}
	}

	// Buat response object dengan semua maklumat yang frontend perlukan
	result := XsollaCallbackResponse{
		Email:        claims.Email,
		Name:         name,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Provider:     claims.Provider,
		UserID:       claims.Sub, // "sub" = subject = unique user ID dalam JWT standard
		Picture:      claims.Picture,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// decodeJWTClaims — decode bahagian payload JWT token tanpa verify signature
// Kita just nak baca maklumat user — signature verification buat dalam middleware
func decodeJWTClaims(token string) (XsollaUserClaims, error) {
	var claims XsollaUserClaims

	// JWT format: "header.payload.signature" — split by "."
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return claims, errors.New("invalid jwt format")
	}

	payload := parts[1] // Bahagian kedua = payload (maklumat user)

	// Payload dikodkan dalam base64 — kita decode balik jadi JSON
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return claims, err
	}

	// Parse JSON untuk ambil field-field dalam XsollaUserClaims
	err = json.Unmarshal(decoded, &claims)
	return claims, err
}
