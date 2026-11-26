package request

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestConfig_BaseURL(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cfg := New(Settings{BaseURL: srv.URL + "/api/v1"})

	if _, err := cfg.Send[struct{}](Options{Method: http.MethodGet, Url: "/users/1"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if gotPath != "/api/v1/users/1" {
		t.Errorf("path = %q, want /api/v1/users/1", gotPath)
	}

	// Absolute URLs bypass the BaseURL.
	if _, err := cfg.Send[struct{}](Options{Method: http.MethodGet, Url: srv.URL + "/other"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if gotPath != "/other" {
		t.Errorf("path = %q, want /other", gotPath)
	}
}

func TestConfig_HeadersMerge(t *testing.T) {
	var gotAPI, gotTrace, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPI = r.Header.Get("X-Api-Key")
		gotTrace = r.Header.Get("X-Trace")
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cfg := New(Settings{
		BaseURL: srv.URL,
		Headers: map[string]string{
			"X-Api-Key":    "cfg-key",
			"X-Trace":      "cfg-trace",
			"Content-Type": "application/json",
		},
	})

	_, err := cfg.Send[struct{}](Options{
		Method: http.MethodPost,
		Url:    "/ping",
		Headers: map[string]string{
			"X-Trace":      "local-trace", // overrides config
			"Content-Type": "text/plain",  // overrides config and automatic
		},
		Body: map[string]any{"a": 1},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if gotAPI != "cfg-key" {
		t.Errorf("X-Api-Key = %q, want cfg-key", gotAPI)
	}
	if gotTrace != "local-trace" {
		t.Errorf("X-Trace = %q, want local-trace", gotTrace)
	}
	if gotCT != "text/plain" {
		t.Errorf("Content-Type = %q, want text/plain", gotCT)
	}
}

func TestConfig_Auth(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cfg := New(Settings{BaseURL: srv.URL, Auth: BearerAuth{Token: "cfg-tok"}})

	// Config auth is applied when Options.Auth is nil.
	if _, err := cfg.Send[struct{}](Options{Method: http.MethodGet, Url: "/a"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got != "Bearer cfg-tok" {
		t.Errorf("Authorization = %q, want Bearer cfg-tok", got)
	}

	// Options.Auth wins over the config one.
	if _, err := cfg.Send[struct{}](Options{Method: http.MethodGet, Url: "/b", Auth: BearerAuth{Token: "local-tok"}}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got != "Bearer local-tok" {
		t.Errorf("Authorization = %q, want Bearer local-tok", got)
	}
}

func TestConfig_ClientUsed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var used atomic.Int32
	cfg := New(Settings{
		BaseURL: srv.URL,
		Client: &http.Client{Transport: markTransport{
			inner:  http.DefaultTransport,
			onTrip: func() { used.Add(1) },
		}},
	})

	if _, err := cfg.Send[struct{}](Options{Method: http.MethodGet, Url: "/x"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if err := cfg.Download(Options{Method: http.MethodGet, Url: "/file"}, t.TempDir()+"/f"); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if used.Load() == 0 {
		t.Error("config client transport was never invoked")
	}

	// Options.Client wins over the config one.
	before := used.Load()
	if _, err := cfg.Send[struct{}](Options{Method: http.MethodGet, Url: "/x", Client: http.DefaultClient}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if used.Load() != before {
		t.Error("Options.Client should bypass the config client, but transport was invoked")
	}
}

type markTransport struct {
	inner  http.RoundTripper
	onTrip func()
}

func (t markTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.onTrip()
	return t.inner.RoundTrip(r)
}

// The headline use case: one config carrying custom TLS certificates is
// enough to talk to a TLS server with a private CA.
func TestConfig_CustomTLS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	cfg := New(Settings{
		BaseURL: srv.URL,
		Client:  &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}},
	})

	resp, err := cfg.Send[struct{ Ok bool }](Options{Method: http.MethodGet, Url: "/ping"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp == nil || !resp.Ok {
		t.Errorf("decoded result = %+v, want {ok:true}", resp)
	}
}

// Regression: Download must honor the config's Client (previously it
// always used a fresh http.Client, silently ignoring custom TLS).
func TestConfig_DownloadUsesTLSClient(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("cert-data"))
	}))
	defer srv.Close()

	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	cfg := New(Settings{
		Client: &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}},
	})

	to := t.TempDir() + "/certs.pem"
	if err := cfg.Download(Options{Method: http.MethodGet, Url: srv.URL}, to); err != nil {
		t.Fatalf("Download: %v", err)
	}
	b, err := os.ReadFile(to)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !strings.Contains(string(b), "cert-data") {
		t.Errorf("file content = %q, want it to contain cert-data", b)
	}
}

// Two independent configs, like two package-level variables for two
// environments (ru / www).
func TestConfig_IndependentInstances(t *testing.T) {
	var ruPath, wwwPath string
	ruSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ruPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ruSrv.Close()
	wwwSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wwwPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer wwwSrv.Close()

	ru := New(Settings{BaseURL: ruSrv.URL + "/ru"})
	www := New(Settings{BaseURL: wwwSrv.URL + "/www"})

	if _, err := ru.Send[struct{}](Options{Method: http.MethodGet, Url: "/users/1"}); err != nil {
		t.Fatalf("ru.Send: %v", err)
	}
	if _, err := www.Send[struct{}](Options{Method: http.MethodGet, Url: "/users/1"}); err != nil {
		t.Fatalf("www.Send: %v", err)
	}
	if ruPath != "/ru/users/1" {
		t.Errorf("ru path = %q, want /ru/users/1", ruPath)
	}
	if wwwPath != "/www/users/1" {
		t.Errorf("www path = %q, want /www/users/1", wwwPath)
	}
}

// Package-level helpers must work with no config at all — zero Settings
// means zero behavior change.
func TestConfig_ZeroSettings_NoBehaviorChange(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"abc"}`))
	}))
	defer srv.Close()

	cfg := New(Settings{})

	resp, err := cfg.Send[YooKassaResponse](Options{
		Method: http.MethodGet,
		Url:    srv.URL,
		Auth:   BearerAuth{Token: "tok"},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp == nil || resp.ID != "abc" {
		t.Errorf("decoded result = %+v, want {abc}", resp)
	}
	if gotAuth != "Bearer tok" {
		t.Errorf("Authorization = %q, want Bearer tok", gotAuth)
	}
}
