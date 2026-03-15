package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/theabgarg/market-aggregator/internal/domain"
)

func TestHandleAggregator(t *testing.T) {
	redisAddr := "localhost:6379" 
	cache := domain.NewMarketCache(redisAddr)
	cache.Update("BTCUSDT", 65000.50)

	h := NewHandler(cache, nil, nil)

	tests := []struct {
		name           string
		symbolQuery    string
		expectedStatus int
		expectedPrice  float64
	}{
		{
			name:           "Valid Symbol Returns Price",
			symbolQuery:    "BTCUSDT",
			expectedStatus: http.StatusOK,
			expectedPrice:  65000.50,
		},
		{
			name:           "Missing Symbol Returns 404",
			symbolQuery:    "UNKNOWN",
			expectedStatus: http.StatusNotFound,
			expectedPrice:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/prices?symbol="+tc.symbolQuery, nil)
			rr := httptest.NewRecorder()

			h.handleAggregator(rr, req)
			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %v, go %v", tc.expectedStatus, rr.Code)
			}

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)

				if err != nil {
					t.Fatalf("Failed to decode response body: %v", err)
				}
				actualPrice, ok := response["price"].(float64)
				if !ok {
					t.Fatalf("price field missing or not a number")
				}

				if actualPrice != tc.expectedPrice {
					t.Errorf("expected price %v, got %v", tc.expectedPrice, actualPrice)
				}
			}
		})
	}
}
