// Package main shows how to GET a JSON resource and decode it into a
// typed struct using request.Send.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lazylib/request"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	u, err := request.Send[User](request.Options{
		Method: http.MethodGet,
		Url:    "https://api.example.com/users/1",
		Auth:   request.BearerAuth{Token: "secret"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("user %d: %s\n", u.ID, u.Name)
}
