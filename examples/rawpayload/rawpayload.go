// Package main shows how to send a pre-built raw payload (e.g. a
// pre-signed webhook body, a file, or a string) with request.Send.
package main

import (
	"log"
	"net/http"

	"github.com/lazylib/request"
)

func main() {
	// []byte is sent as-is, no JSON re-encoding, no automatic Content-Type.
	// Supply your own Content-Type via Headers if the server needs it.
	_, err := request.Send[struct{}](request.Options{
		Method: http.MethodPost,
		Url:    "https://api.example.com/hooks/42",
		Body:   []byte(`{"event":"ping"}`),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Auth: request.BearerAuth{Token: "secret"},
	})
	if err != nil {
		log.Fatal(err)
	}
}
