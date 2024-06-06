package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
)

// Token struct to hold the token data
type Token struct {
	Token   string `json:"token"`
	Expiry  int64  `json:"expiry"`
	Created int64  `json:"created"`
}

var GlobalToken Token
var prevURL string

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading environment variables")
	}
}

func main() {

	accessToken := os.Getenv("TW_ACCESS_TOKEN")
	accessSecret := os.Getenv("TW_ACCESS_SECRET")
	if accessToken == "" || accessSecret == "" {
		fmt.Fprintln(os.Stderr, "Please set the TW_ACCESS_TOKEN and TW_ACCESS_SECRET environment variables.")
		os.Exit(1)
	}

	var token, err = GetTokenFromAirtable()
	if err != nil {
		fmt.Println("Unable to get token from airtable: ", err)
	}
	GlobalToken = token
	if GlobalToken.Token == "" {
		UpdateToken()
	}
	if !TokenValid(GlobalToken) {
		UpdateToken()
	}

	url, err := GetCurrentlyPlaying(GlobalToken.Token)
	if err != nil {
		fmt.Println("Error getting currently playing track from spotify: ", err)
	}

	if url == prevURL {
		return
	}

	client, err := newOAuth1Client(accessToken, accessSecret)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = tweet(client, url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	prevURL = url
}

func GetNewToken() (*http.Response, error) {
	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", os.Getenv("REFRESH_TOKEN"))

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	encodedSecret := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))

	headers := http.Header{}
	headers.Set("Authorization", "Basic "+encodedSecret)
	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header = headers

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func UpdateToken() error {
	//get new token from spotify
	resp, err := GetNewToken()
	if err != nil {
		return fmt.Errorf("error fetching new token: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	// Parse the JSON response
	var resData struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &resData); err != nil {
		return fmt.Errorf("error parsing response JSON: %v", err)
	}

	// Spotify tokens expire after 1 hour. Convert expiry time to milliseconds and take 300000ms off to account for any latency
	created := time.Now().Unix() * 1000
	token := Token{
		Token:   resData.AccessToken,
		Expiry:  (resData.ExpiresIn - 300) * 1000,
		Created: created,
	}

	// Update global token variable
	GlobalToken = token

	if err := UpdateAirtable(token); err != nil {
		return fmt.Errorf("error updating Airtable: %v", err)
	}

	return nil
}

func UpdateAirtable(token Token) error {
	apiKey := os.Getenv("AIRTABLE_API_KEY")
	baseID := os.Getenv("AIRTABLE_BASE_ID")
	tableID := os.Getenv("AIRTABLE_TABLE_ID")
	recordID := os.Getenv("RECORD_ID")

	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"token":   token.Token,
			"expiry":  token.Expiry,
			"created": token.Created,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error encoding payload to JSON: %v", err)
	}

	//create new PUT request
	req, err := http.NewRequest("PATCH", fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s", baseID, tableID, recordID), strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error updating Airtable: received status code %d", resp.StatusCode)
	}

	return nil
}

func TokenValid(token Token) bool {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	expiry := token.Created + token.Expiry

	return now < expiry
}

func GetCurrentlyPlaying(token string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	if err != nil {
		return "", fmt.Errorf("error creating request to get currently playing: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusNoContent {
		os.Exit(0)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error getting current track: received status code %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	item, ok := responseBody["item"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("item not found in response")
	}

	externalURLs, ok := item["external_urls"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("external_urls not found in item")
	}

	spotifyURL, ok := externalURLs["spotify"].(string)
	if !ok {
		return "", fmt.Errorf("spotify URL not found in external_urls")
	}
	return spotifyURL, nil
}

func GetTokenFromAirtable() (Token, error) {
	apiKey := os.Getenv("AIRTABLE_API_KEY")
	baseID := os.Getenv("AIRTABLE_BASE_ID")
	tableID := os.Getenv("AIRTABLE_TABLE_ID")
	recordID := os.Getenv("RECORD_ID")

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s", baseID, tableID, recordID), nil)

	if err != nil {
		return Token{}, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("error fetching token from airtable: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Token{}, fmt.Errorf("error reading response body: %v", err)
	}

	// Parse the JSON response
	var airtableToken Token
	if err := json.Unmarshal(body, &airtableToken); err != nil {
		return Token{}, fmt.Errorf("error parsing response JSON: %v", err)
	}

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("error getting token: received status code %d", resp.StatusCode)
	}

	return airtableToken, nil
}

func newOAuth1Client(accessToken, accessSecret string) (*gotwi.Client, error) {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           accessToken,
		OAuthTokenSecret:     accessSecret,
	}

	return gotwi.NewClient(in)
}

func tweet(c *gotwi.Client, text string) error {
	p := &types.CreateInput{
		Text: gotwi.String(text),
	}

	_, err := managetweet.Create(context.Background(), c, p)
	if err != nil {
		return err
	}

	return nil
}
