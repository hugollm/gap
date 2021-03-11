# Panic

If an endpoint panics, the framework will send a generic server error response:

```
500 Internal Server Error

{"error": "internal server error"}
```

## Error Handler

By default, panics are recovered by the app, logged to stdout and the 500 response is sent. This is the default error handler:

```go
func defaultErrorHandler(ierr interface{}, response http.ResponseWriter) {
    log.Print("PANIC: ", ierr)
    response.Header().Set("Content-Type", "application/json")
    response.WriteHeader(500)
    response.Write([]byte(`{"error": "internal server error"}\n`))
}
```

One can override this behavior by replacing the error handler on the `App`:

```go
app.ErrorHandler(myErrorHandler)
```

As with the default handler, the custom error handler is a function that must take the error interface and response writer. Here's an example:

```go
import "net/http"

func myErrorHandler(ierr interface{}, response http.ResponseWriter) {
    // send metrics somewhere...
    response.WriteHeader(503)
    response.Write([]byte(`Service Unavailable`))
}
```


## Panic Custom Error

It's also possible to panic a custom error that binds to response. This is a shortcut for aborting requests, usually most useful on reusable logic.

Let's say you want to craft a reusable `auth` function, to reuse across endpoints. This panic mechanism will allow you to send a 401 response from this function, instead of repeating the `if err` logic every time. e.g.

```go
type authError struct {
    Status int `response:"status"`
    Message string `response:"json,auth_error"`
}

func (err authError) Error() string {
    return err.Message
}

func auth(token string) User {
    // ...
    if err != nil {
        panic(authError{401, "missing or invalid access token"})
    }
    return token
}
```

Endpoints can then make use of this logic without worrying about auth error responses:

```
func myProfile(input myProfileInput) myProfileOutput {
    user := auth(input.Token)
    // ...
}
```
