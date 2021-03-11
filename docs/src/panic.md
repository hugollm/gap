# Panic

If an endpoint panics, the framework will send a generic server error response:

```
500 Internal Server Error

{"error": "internal server error"}
```

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
