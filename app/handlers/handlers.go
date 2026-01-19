package handlers

import (
	"fmt"
	"log"

	"faucet/app/config"
	"faucet/app/lightning"
	"faucet/app/security"

	"github.com/gofiber/fiber/v2"
)

func HandleIndex(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"site_key":   config.SiteKey,
		"max_amount": config.MaxClaimAmount,
	})
}

func HandleBalance(c *fiber.Ctx) error {
	balance, err := lightning.FetchBalance()
	if err != nil {
		log.Printf("Balance Fetch Error: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"balance": 0,
			"error":   "Could not connect to wallet",
		})
	}
	return c.JSON(fiber.Map{"balance": balance})
}

func HandleClaim(c *fiber.Ctx, tracker *security.RateLimiter) error {
	type ClaimRequest struct {
		CaptchaToken string `json:"captcha_token"`
		Invoice      string `json:"invoice"`
	}

	var payload ClaimRequest

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": "Invalid JSON"})
	}

	// amount validation
	amounts, err := lightning.GetLNInvoiceAmount(payload.Invoice)
	fmt.Printf("Amounts obj: %+v\n", amounts)

	if err != nil || amounts.Sats > config.MaxClaimAmount || amounts.Sats < 1 {
		return c.Status(400).JSON(fiber.Map{
			"detail": fmt.Sprintf("Invalid amount. Max is %d sats.", config.MaxClaimAmount),
		})
	}

	// rate limit
	clientIP := c.Get("X-Real-IP", c.IP())
	allowed, waitTime := tracker.IsAllowed(clientIP)
	if !allowed {
		minutes := int(waitTime.Minutes())
		return c.Status(429).JSON(fiber.Map{
			"detail": fmt.Sprintf("Rate limit exceeded. Please wait %d minutes.", minutes),
		})
	}

	// captcha
	if !security.VerifyTurnstile(payload.CaptchaToken, clientIP) {
		return c.Status(400).JSON(fiber.Map{"detail": "Captcha failed"})
	}

	// invoice payment
	hash, err := lightning.PayInvoice(payload.Invoice)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": "Payment failed: " + err.Error()})
	}

	tracker.RecordClaim(clientIP)

	return c.JSON(fiber.Map{
		"paymentHash": hash,
		"status":      "success",
	})
}