package request

import (
	"net/http"
	"strings"
)

// Settings describes the defaults applied to every request made through a
// [Config] created with [New].
//
// The typical use case is talking to an API that needs custom TLS
// certificates and constant authorization: configure the client and auth
// once, then keep call sites down to Method/Url/Body.
type Settings struct {
	// BaseURL is prepended to relative Url values in [Options].
	//
	//	Settings{BaseURL: "https://api.example.com/v1"}
	//	Options{Url: "/users/1"} // -> https://api.example.com/v1/users/1
	//
	// Absolute URLs (containing "://") are used as-is, so a single call
	// can hit another host without touching the config.
	BaseURL string

	// Headers are added to every request (e.g. X-Api-Key, Accept).
	// Headers from Options override them; both override the automatic
	// Content-Type set for JSON bodies.
	Headers map[string]string

	// Auth is the default authorization: [BasicAuth], [BearerAuth] or a
	// custom [Auth] implementation.
	Auth Auth

	// Client is the http.Client used for every request. Custom TLS
	// certificates, timeouts and proxies are configured here.
	Client *http.Client
}

// Config is a reusable client bound to a fixed [Settings]. Create one with
// [New], keep it in your own package-level variable and use it anywhere:
//
//	var api = request.New(request.Settings{
//	    BaseURL: "https://api.example.com/v1",
//	    Auth:    request.BearerAuth{Token: token},
//	})
//
//	user, err := api.Send[User](request.Options{Method: http.MethodGet, Url: "/users/42"})
//
// Create as many independent configs as you need (one per API, per
// environment, per region).
//
// Config is a value type and safe for concurrent use. Treat the fields of
// its Settings (Headers map, Auth, Client) as read-only after creation.
type Config struct {
	settings Settings
}

// New builds a Config that applies settings to every request made through
// it. New never fails: all validation happens per request, exactly like
// with the package-level helpers.
func New(settings Settings) Config {
	return Config{settings: settings}
}

// Send performs opts and decodes the response body into *T, applying the
// config's settings first. Otherwise identical to the package-level [Send].
func (c Config) Send[T any](opts Options) (*T, error) {
	return Send[T](c.resolve(opts))
}

// SendX is the panicking variant of [Config.Send].
// Otherwise identical to the package-level [SendX].
func (c Config) SendX[T any](opts Options) *T {
	return SendX[T](c.resolve(opts))
}

// Download performs opts and writes the response body to the file path
// specified by to, applying the config's settings first.
// Otherwise identical to the package-level [Download].
func (c Config) Download(opts Options, to string) error {
	return Download(c.resolve(opts), to)
}

// DownloadX performs the same operation as [Config.Download] but panics
// on error. Otherwise identical to the package-level [DownloadX].
func (c Config) DownloadX(opts Options, to string) {
	DownloadX(c.resolve(opts), to)
}

// resolve returns opts with the config's settings applied.
// Values set directly in opts always win over the config ones.
func (c Config) resolve(opts Options) Options {
	s := c.settings

	if opts.Auth == nil {
		opts.Auth = s.Auth
	}
	if opts.Client == nil {
		opts.Client = s.Client
	}
	if opts.Url != "" && s.BaseURL != "" && !isAbsoluteURL(opts.Url) {
		opts.Url = joinURL(s.BaseURL, opts.Url)
	}
	if len(s.Headers) > 0 {
		merged := make(map[string]string, len(s.Headers)+len(opts.Headers))
		for k, v := range s.Headers {
			merged[k] = v
		}
		for k, v := range opts.Headers {
			merged[k] = v
		}
		opts.Headers = merged
	}
	return opts
}

// isAbsoluteURL reports whether url carries a scheme ("://").
func isAbsoluteURL(url string) bool {
	return strings.Contains(url, "://")
}

// joinURL concatenates base and path without duplicating slashes:
//
//	joinURL("https://api.example.com/v1", "/users/1")
//	// -> "https://api.example.com/v1/users/1"
func joinURL(base, path string) string {
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimLeft(path, "/")
}
