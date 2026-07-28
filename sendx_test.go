package request

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendX_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"abc","status":"pending"}`))
	}))
	defer srv.Close()

	got := SendX[YooKassaResponse](Options{Method: http.MethodGet, Url: srv.URL})
	if got == nil || got.ID != "abc" || got.Status != "pending" {
		t.Errorf("SendX result = %+v, want {abc pending}", got)
	}
}

func TestSendX_PanicsOnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got nil")
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("recovered value is not an error: %T", r)
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("panic value = %v, want it to mention 500", err)
		}
	}()

	_ = SendX[YooKassaResponse](Options{Method: http.MethodGet, Url: srv.URL})
}
