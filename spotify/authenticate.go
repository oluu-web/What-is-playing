package spotify

import (
	"fmt"
	"net/http"
	"os"
)

// Authenticate redirects users to Spotify's authorization endpoint
func Authenticate(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	fmt.Println("this!!! --->!", clientID)
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	authEndpoint := "https://accounts.spotify.com/authorize"

	authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=user-read-currently-playing", authEndpoint, clientID, redirectURI)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
