# Error

Endpoints can optionally return an `error`. It can be any valid go `error`. Example:

```go
import "errors"

func auth() error {
    // ...
    return errors.New("invalid access token")
}
```

This will result in a "bad request" response:

```
400 Bad Request

{"error": "invalid access token"}
```

The status is 400 and the message is extracted from `err.Error()`.


Although this might be good enough for simple use cases, sometimes it's necessary to send more complex error responses. For those cases, you can create a custom `error` from a struct, that binds to a response exactly how the output struct does.

Here's a simple example:

```go
type authError struct {
    Status int `response:"status"`
    Message string `response:"json,auth_error"`
}

func (err authError) Error() string {
    return err.Message
}
```

Note how you need to implement the `Error() string` method even if you don't plan to use it. This is required so your struct is recognized as an `error` on go.

An endpoint can then make use of this custom error:

```go
func auth() error {
    // ...
    return authError{401, "invalid access token"}
}
```

This will then translate to the response:

```
401 Unauthorized

{"auth_error": "invalid access token"}
```
