package security

import (
	"encoding/json"
	"faucet/app/config"
	"net/http"
	"net/url"
	"time"
)

func VerifyTurnstile(token, ip string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	data := url.Values{
		"secret":   {config.TurnstileSecret},
		"response": {token},
		"remoteip": {ip},
	}

	resp, err := client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", data)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Success
}