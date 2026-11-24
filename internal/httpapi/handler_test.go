package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalculateEndpoint(t *testing.T) {
	const valid = `{"expression":"0-3"}`
	tests := []struct {
		name, method, contentType, body string
		status                          int
	}{
		{"valid zero operand", "POST", "application/json; charset=utf-8", valid, 200},
		{"wrong method", "GET", "", "", 405},
		{"wrong content type", "POST", "text/plain", valid, 415},
		{"malformed JSON", "POST", "application/json", `{`, 400},
		{"missing expression", "POST", "application/json", `{}`, 400},
		{"empty expression", "POST", "application/json", `{"expression":""}`, 400},
		{"malformed expression", "POST", "application/json", `{"expression":"1+"}`, 400},
		{"consecutive signs", "POST", "application/json", `{"expression":"1+2*3-+2^2"}`, 400},
		{"unknown field", "POST", "application/json", `{"expression":"1+2","extra":true}`, 400},
		{"legacy binary payload", "POST", "application/json", `{"operation":"add","a":1,"b":2}`, 400},
		{"legacy arrays payload", "POST", "application/json", `{"numbers":[1,2],"operators":["+"]}`, 400},
		{"trailing JSON", "POST", "application/json", valid + `{}`, 400},
		{"oversized body", "POST", "application/json", `{"expression":"1+2"}` + strings.Repeat(" ", maxBodyBytes), 413},
		{"unknown operator", "POST", "application/json", `{"expression":"1%2"}`, 400},
		{"division by zero", "POST", "application/json", `{"expression":"1/0"}`, 422},
		{"number exceeds supported range", "POST", "application/json", `{"expression":"270000000000000000000+20"}`, 422},
	}

	handler := NewHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/calculate", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.status, rec.Body)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q", got)
			}
			if rec.Code == http.StatusMethodNotAllowed && rec.Header().Get("Allow") != "POST" {
				t.Fatal("expected Allow: POST")
			}
			if rec.Code == http.StatusOK {
				var response calculateResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Result != -3 {
					t.Fatalf("unexpected success body: %s", rec.Body)
				}
			} else {
				var response errorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Error == "" {
					t.Fatalf("unexpected error body: %s", rec.Body)
				}
			}
		})
	}
}

func TestCalculateInfixExpressionEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"expression":"1+2*3-2^2"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)

	var response calculateResponse
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &response) != nil || response.Result != 3 {
		t.Fatalf("unexpected response: status=%d body=%s", rec.Code, rec.Body)
	}
}

func TestCalculateDecimalAdditionEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"expression":"27.02+20"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)

	var response calculateResponse
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &response) != nil || response.Result != 47.02 {
		t.Fatalf("unexpected response: status=%d body=%s", rec.Code, rec.Body)
	}
}

func TestCalculateLargeAdditionEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"expression":"2700000000000000+20"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)

	var response calculateResponse
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &response) != nil || response.Result != 2700000000000020 {
		t.Fatalf("unexpected response: status=%d body=%s", rec.Code, rec.Body)
	}
}

func TestCalculateInfixExpressionValidation(t *testing.T) {
	tests := []struct {
		expression string
		status     int
		result     float64
	}{
		{"1+4+3", http.StatusOK, 8},
		{"1+4+3-", http.StatusBadRequest, 0},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"expression":"`+tt.expression+`"}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			NewHandler().ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.status, rec.Body)
			}
			if tt.status == http.StatusOK {
				var response calculateResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Result != tt.result {
					t.Fatalf("unexpected success body: %s", rec.Body)
				}
			}
		})
	}
}
