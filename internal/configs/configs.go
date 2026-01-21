package configs

import (
	"os"
	"strconv"
	"strings"
)

var (
	LnbitsUrl       string
	LnbitsKey       string
	TurnstileSecret string
	SiteKey         string
	MaxClaimAmount  int
	RateLimitWindow int
)

func LoadConfig() {
	LnbitsUrl = strings.TrimSuffix(os.Getenv("LNBITS_URL"), "/")
	LnbitsKey = os.Getenv("LNBITS_ADMIN_KEY")
	TurnstileSecret = os.Getenv("TURNSTILE_SECRET")
	SiteKey = os.Getenv("TURNSTILE_SITE_KEY")

	if LnbitsKey == "" || TurnstileSecret == "" {
		panic("Missing critical environment variables! Check your .env file.")
	}

	MaxClaimAmount, _ = strconv.Atoi(os.Getenv("MAX_CLAIM_AMOUNT"))
	if MaxClaimAmount == 0 {
		MaxClaimAmount = 100
	}

	RateLimitWindow, _ = strconv.Atoi(os.Getenv("RATE_LIMIT_WINDOW"))
	if RateLimitWindow == 0 {
		RateLimitWindow = 3600
	}
}