// Package main shows how to send a request with HTTP Basic auth
// (e.g. payment gateways, internal services, S3-compatible APIs) using
// request.Send.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lazylib/request"
)

type Payment struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func main() {
	p, err := request.Send[Payment](request.Options{
		Method: http.MethodPost,
		Url:    "https://api.yookassa.ru/v3/payments",
		Body: map[string]any{
			"amount":  map[string]string{"value": "100.00", "currency": "RUB"},
			"capture": true,
		},
		Auth: &request.BasicAuth{Username: "shop-id", Password: "secret"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("payment %s: %s\n", p.ID, p.Status)
}
