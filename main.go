package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/oluu-web/what-is-playing/spotify"
)

func main() {
	// Start the authentication flow when the user navigates to /spotify/auth
	http.HandleFunc("/spotify/auth", spotify.Authenticate)

	// Handle the callback after Spotify authentication
	http.HandleFunc("/callback", handleCallback)

	// Start the server
	log.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	// Extract the authorization code from the callback URL
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// Exchange the code for an access token
	accessToken, err := spotify.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange code for token: %s", err), http.StatusInternalServerError)
		return
	}

	// Use the access token to fetch currently playing track
	currentTrack, err := spotify.FetchCurrentSong(accessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch currently playing track: %s", err), http.StatusInternalServerError)
		return
	}

	// Display the currently playing track information
	// fmt.Fprintf(w, "Currently playing track: %s by %s from album %s", currentTrack.Name, currentTrack.Artist, currentTrack.Album)
	fmt.Println(currentTrack)
}
