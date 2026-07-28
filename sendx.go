package request

// SendX is the panicking variant of [Send].
//
// It calls Send[T] and panics with the returned error if it is non-nil.
// It is meant for cases where an HTTP failure is a programmer error or
// a process-fatal condition (similar to entgo.io/ent's Must helpers, or
// regexp.MustCompile): the request is so fundamental to the program's
// flow that handling the error would only obscure control flow.
//
// For most code paths, prefer [Send] and return the error.
//
// Example:
//
//	user := request.SendX[User](request.Options{
//	    Method: http.MethodGet,
//	    Url:    "https://api.example.com/users/1",
//	    Auth:   request.BearerAuth{Token: "secret"},
//	})
//	fmt.Println(user.Name) // no error to check
func SendX[T any](opts Options) *T {
	resp, err := Send[T](opts)
	if err != nil {
		panic(err)
	}
	return resp
}
