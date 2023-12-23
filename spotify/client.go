package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	apiBaseURL = "https://api.spotify.com/v1/"
)

type CurrentlyPlayingTrack struct {
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration_ms"`
	// Include other relevant fields from the API response
}

func FetchCurrentSong(accessToken string) (*CurrentlyPlayingTrack, error) {
	endpoint := apiBaseURL + "me/player/currently-playing"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch currently playing track: %s", resp.Status)
	}
	var currentlyPlaying CurrentlyPlayingTrack
	err = json.NewDecoder(resp.Body).Decode(&currentlyPlaying)
	if err != nil {
		return nil, err
	}
	fmt.Println(currentlyPlaying)

	return &currentlyPlaying, nil
}
