package lightning

import (
	"encoding/json"
	"faucet/app/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchBalance(t *testing.T) {
	const testApiKey string = "test-api-key"

	// mocked server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") !=  testApiKey {
			t.Errorf("Expected X-Api-Key '%s', got '%s'", testApiKey, r.Header.Get("X-Api-Key"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"balance": 50})
	}))
	defer server.Close()

	// mocked server config
	config.LnbitsUrl = server.URL
	config.LnbitsKey = testApiKey

	balance, err := FetchBalance()

	if err != nil {
		t.Fatalf("FetchBalance failed: %v", err)
	}
	if balance != 50 {
		t.Errorf("Expected balance 50, got %d", balance)
	}
}

func TestGetLNInvoiceAmount(t *testing.T) {
	tests := []struct {
		invoice  string
		expSatsValue int
	}{
		{"lnbc100n1...", 10},  		// 100n = 10 sats
		{"lnbc1u1...", 100},   		// 1u = 100 sats
		{"lnbc10m1...", 1000000},	// 10m = 1M sats
	}

	for _, tt := range tests {
		amt, err := GetLNInvoiceAmount(tt.invoice)
		if err != nil {
			t.Errorf("Invoice %s failed: %v", tt.invoice, err)
			continue
		}
		if amt.Sats != tt.expSatsValue {
			t.Errorf("For %s expected %d, got %d (sats)", tt.invoice, tt.expSatsValue, amt.Sats)
		}
	}
}