package spotify

import (
	"net/http"
)

// Authenticate redirects users to Spotify's authorization endpoint
func Authenticate(w http.ResponseWriter, r *http.Request) {
	// clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	// fmt.Println("this!!! --->!", clientID)
	// redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	// authEndpoint := "https://accounts.spotify.com/authorize"

	// authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=user-read-currently-playing", authEndpoint, clientID, redirectURI)
	authURL := "https://accounts.spotify.com/authorize?client_id=0486437cb9a64916849aefdb00c674fd&response_type=code&redirect_uri=http://localhost:8080/callback&scope=user-read-currently-playing"
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
