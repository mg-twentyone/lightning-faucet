package lightning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strings"
	"strconv"
	"faucet/app/config"
)

type Amount struct {
    Sats int `json:"sats"`
    Btc  float64 `json:"btc"`
    Unit string `json:"unit"`
}

func FetchBalance() (int, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", config.LnbitsUrl+"/api/v1/wallet", nil)
	req.Header.Set("X-Api-Key", config.LnbitsKey)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data struct {
		Balance int `json:"balance"` // LNbits returns msats
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	return data.Balance / 1000, nil
}

func PayInvoice(bolt11 string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"out":    true,
		"bolt11": bolt11,
	})

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("POST", config.LnbitsUrl+"/api/v1/payments", bytes.NewBuffer(body))
	req.Header.Set("X-Api-Key", config.LnbitsKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("LNbits error: status %d", resp.StatusCode)
	}

	var result struct {
		PaymentHash string `json:"payment_hash"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.PaymentHash, nil
}

func CreateWithdrawLink(minAmount, maxAmount int) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"title":            fmt.Sprintf("Faucet %d-%d sats", minAmount, maxAmount),
		"min_withdrawable": minAmount,
		"max_withdrawable": maxAmount,
		"uses":             1,
		"wait_time":        1,
		"is_unique":        true,
	})

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("POST", config.LnbitsUrl+"/withdraw/api/v1/links", bytes.NewBuffer(body))
	req.Header.Set("X-Api-Key", config.LnbitsKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		LNURL string `json:"lnurl"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.LNURL, nil
}

func GetLNInvoiceAmount(invoice string) (Amount, error) {
	invoice = strings.ToLower(invoice)
	if !strings.HasPrefix(invoice, "ln") {
		return Amount{}, fmt.Errorf("invalid lightning invoice prefix")
	}

	amountIndex := strings.IndexAny(invoice, "0123456789")
	if amountIndex == -1 {
		return Amount{}, fmt.Errorf("no amount found in the invoice")
	}

	var amountStr strings.Builder
	var multiplier rune
	dataPart := invoice[amountIndex:]

	for _, char := range dataPart {
		if char >= '0' && char <= '9' {
			amountStr.WriteRune(char)
		} else {
			multiplier = char
			break
		}
	}

	amount, err := strconv.ParseFloat(amountStr.String(), 64)
	if err != nil { return Amount{}, err }

	var btc, sats float64
	switch multiplier {
		case 'm': btc = amount * 0.001
		case 'u': btc = amount * 0.000001
		case 'n': btc = amount * 0.000000001
		case 'p': btc = amount * 0.000000000001
		default:  btc = amount * 0.00000001
	}

	sats = btc *  100_000_000

	return Amount{
		Sats: int(sats),
		Btc: btc,
		Unit: string(multiplier),
	}, nil
}