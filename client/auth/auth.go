package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"harmony/client/clip"
	"harmony/client/common"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
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
	Interval        int    `json:"interval"`
}

type Email struct {
	Email      string `json:"email"`
	Verified   bool   `json:"verified"`
	Primary    bool   `json:"primary"`
	Visibility string `json:"visibility"`
}

func requestDeviceCode() (*DeviceAuthResponse, error) {
	data := fmt.Sprintf("client_id=%s&scope=user", clientID)
	req, err := http.NewRequest("POST", deviceAuthURL, bytes.NewBufferString(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := common.Client.Do(req)
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

func pollForToken(deviceCode string) (string, error) {
	for {
		data := fmt.Sprintf("client_id=%s&device_code=%s&grant_type=urn:ietf:params:oauth:grant-type:device_code",
			clientID, deviceCode)
		req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data))
		if err != nil {
			return "", err
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := common.Client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		var tokenResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		if err != nil {
			return "", err
		}

		if _, exists := tokenResp["error"]; exists {
			codes := []string{"authorization_pending", "slow_down"}
			if slices.Contains(codes, tokenResp["error"].(string)) {
				time.Sleep(pollInterval)
				continue
			} else {
				return "", fmt.Errorf("[error] %v", tokenResp)
			}
		}

		return tokenResp["access_token"].(string), nil
	}
}

func getAccessToken() (string, error) {
	deviceResp, err := requestDeviceCode()
	if err != nil {
		log.Println("Failed to request device code:", err)
		return "", err
	}

	clip.CopyToClipboard([]byte(deviceResp.UserCode), clip.TextType)
	fmt.Printf("Visit: %s\nEnter this code: %s\n[Code has been copied to clipboard ✅]\n", deviceResp.VerificationURI, deviceResp.UserCode)

	token, err := pollForToken(deviceResp.DeviceCode)
	if err != nil {
		log.Fatal("Failed to retrieve access token:", err)
		return "", err
	}

	return token, nil
}

func GetEmail() (string, error) {
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

	resp, err := common.Client.Do(req)
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

func SignIn() error {
	// email, err := GetEmail()
	// if err != nil {
	// 	return err
	// }
	email := "210050002@iitb.ac.in"

	req, err := http.NewRequest("GET", common.Host+"/user?email="+email, nil)
	if err != nil {
		return err
	}

	res, err := common.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func SaveCookies() error {
	file, err := os.Create("cookies.json")
	if err != nil {
		return err
	}
	defer file.Close()

	url, _ := url.Parse(common.Host)
	cookies := common.Client.Jar.Cookies(url)
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(cookies); err != nil {
		return err
	}

	return nil
}

func CreateOrRestoreCookies() (bool, error) {
	file, err := os.Open("cookies.json")
	if err != nil {
		file, err = os.Create("cookies.json")
		if err != nil {
			return false, err
		}
		defer file.Close()
		return false, err
	}
	defer file.Close()

	var cookies []*http.Cookie
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cookies); err != nil {
		return true, err
	}

	url, _ := url.Parse(common.Host)
	for _, cookie := range cookies {
		common.Client.Jar.SetCookies(url, []*http.Cookie{cookie})
	}

	return true, nil
}
