package request

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type YooKassaResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func TestSend_YooKassaStyle(t *testing.T) {
	var (
		gotMethod      string
		gotAuth        string
		gotContentType string
		gotIdempotence string
		gotBody        map[string]any
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotIdempotence = r.Header.Get("Idempotence-Key")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(YooKassaResponse{ID: "abc", Status: "pending"})
	}))
	defer srv.Close()

	body := map[string]any{
		"amount": map[string]string{"value": "100.00", "currency": "RUB"},
	}

	result, err := Send[YooKassaResponse](Options{
		Method: http.MethodPost,
		Url:    srv.URL,
		Body:   body,
		Headers: map[string]string{
			"Content-Type":    "application/json",
			"Idempotence-Key": "idem-1",
		},
		Auth: &BasicAuth{
			Username: "shop-id",
			Password: "secret",
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotIdempotence != "idem-1" {
		t.Errorf("Idempotence-Key = %q, want idem-1", gotIdempotence)
	}
	if !strings.HasPrefix(gotAuth, "Basic ") {
		t.Errorf("Authorization = %q, want Basic prefix", gotAuth)
	}
	if gotBody["amount"].(map[string]any)["value"] != "100.00" {
		t.Errorf("server-side decoded body = %+v, want amount.value=100.00", gotBody)
	}
	if result == nil || result.ID != "abc" || result.Status != "pending" {
		t.Errorf("decoded result = %+v, want {abc pending}", result)
	}
}

func TestSend_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := Send[YooKassaResponse](Options{Method: http.MethodGet, Url: srv.URL})
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %v, want it to mention 500", err)
	}
}

func TestSend_RawBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if string(b) != "raw-payload" {
			t.Errorf("body = %q, want raw-payload", string(b))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	_, err := Send[struct{}](Options{
		Method: http.MethodPost,
		Url:    srv.URL,
		Body:   []byte("raw-payload"),
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestSend_BearerAuth(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if _, err := Send[map[string]any](Options{
		Method: http.MethodGet,
		Url:    srv.URL,
		Auth:   BearerAuth{Token: "tok"},
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got != "Bearer tok" {
		t.Errorf("Authorization = %q, want Bearer tok", got)
	}
}
