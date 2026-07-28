// Package main shows how to send a request with a Bearer token using
// request.Send.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lazylib/request"
)

type Me struct {
	Login string `json:"login"`
}

func main() {
	me, err := request.Send[Me](request.Options{
		Method: http.MethodGet,
		Url:    "https://api.github.com/user",
		Auth:   request.BearerAuth{Token: "ghp_…"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("hello,", me.Login)
}
