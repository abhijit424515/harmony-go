package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

const (
	clientID       = "Iv23lixNQQwpjTJDJGf5"
	deviceAuthURL  = "https://github.com/login/device/code"
	githubEmailAPI = "https://api.github.com/user/emails"
	pollInterval   = 5 * time.Second
	tokenURL       = "https://github.com/login/oauth/access_token"
)

type DeviceAuthResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval,omitempty"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type Email struct {
	Email      string `json:"email"`
	Verified   bool   `json:"verified"`
	Primary    bool   `json:"primary"`
	Visibility string `json:"visibility"`
}

func requestDeviceCode() (*DeviceAuthResponse, error) {
	data := fmt.Sprintf("client_id=%s&scope=repo", clientID)
	resp, err := http.Post(deviceAuthURL, "application/x-www-form-urlencoded", bytes.NewBufferString(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var deviceResp DeviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, err
	}
	return &deviceResp, nil
}

func pollForToken(deviceCode string) (*TokenResponse, error) {
	for {
		data := fmt.Sprintf("client_id=%s&device_code=%s&grant_type=urn:ietf:params:oauth:grant-type:device_code",
			clientID, deviceCode)
		resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", bytes.NewBufferString(data))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusOK {
			var tokenResp TokenResponse
			if err := json.Unmarshal(body, &tokenResp); err != nil {
				return nil, err
			}
			return &tokenResp, nil
		}

		var respData map[string]interface{}
		json.Unmarshal(body, &respData)
		if errorCode, exists := respData["error"]; exists {
			if errorCode == "authorization_pending" {
				time.Sleep(pollInterval)
				continue
			} else { // to be dealt with later
				return nil, fmt.Errorf("error: %v", respData)
			}
		}
	}
}

func getAccessToken() (string, error) {
	deviceResp, err := requestDeviceCode()
	if err != nil {
		log.Fatal("Failed to request device code:", err)
		return "", err
	}

	fmt.Printf("Visit: %s\nEnter this code: %s\n", deviceResp.VerificationURI, deviceResp.UserCode)

	tokenResp, err := pollForToken(deviceResp.DeviceCode)
	if err != nil {
		log.Fatal("Failed to retrieve access token:", err)
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func getEmail() (string, error) {
	token, err := getAccessToken()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", githubEmailAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []Email
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	if len(emails) == 0 {
		return "", fmt.Errorf("no emails found")
	}

	sort.Slice(emails, func(i, j int) bool {
		return emails[i].Email < emails[j].Email
	})

	return emails[0].Email, nil
}
