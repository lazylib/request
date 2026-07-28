// Package main shows how to POST a struct as JSON and decode the
// JSON response with request.Send.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lazylib/request"
)

type CreateUserReq struct {
	Name string `json:"name"`
}

type CreateUserResp struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	resp, err := request.Send[CreateUserResp](request.Options{
		Method: http.MethodPost,
		Url:    "https://api.example.com/users",
		Body:   CreateUserReq{Name: "Alice"},
		Auth:   request.BearerAuth{Token: "secret"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created user %d: %s\n", resp.ID, resp.Name)
}
