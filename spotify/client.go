package spotify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
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

func makeRequestAndLogDetails(req *http.Request) (*http.Response, error) {
	// Log the request details before making the request
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Println("Failed to dump request:", err)
	} else {
		log.Println("Request:")
		log.Println(string(dump))
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Log the response details after receiving the response
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Println("Failed to dump response:", err)
	} else {
		log.Println("Response:")
		log.Println(string(respDump))
	}

	return resp, nil
}

func FetchCurrentSong(accessToken string) (string, error) {
	endpoint := apiBaseURL + "me/player/currently-playing"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	// client := &http.Client{}
	resp, err := makeRequestAndLogDetails(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch currently playing track: %s", resp.Status)
	}
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	context, ok := data["context"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("context field not found in the response")
	}

	externalURL, ok := context["external_urls"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("external_url field not found in the response")
	}

	spotifyURL, ok := externalURL["spotify"].(string)
	if !ok {
		return "", fmt.Errorf("spotify field not found in the response")
	}

	return spotifyURL, nil
}
