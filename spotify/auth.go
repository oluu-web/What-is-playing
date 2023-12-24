package spotify

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

const tokenEndpoint = "https://accounts.spotify.com/api/token"

// ExchangeCodeForToken exchanges the authorization code for an access token
func ExchangeCodeForToken(code string) (string, error) {
	// clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	// clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	// redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	clientID := "0486437cb9a64916849aefdb00c674fd"
	clientSecret := "09966258ba524087a48a8ca600a15f73"
	redirectURI := "http://localhost:8080/callback"

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}
