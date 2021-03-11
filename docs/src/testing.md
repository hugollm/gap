# Testing

Testing endpoints is quite easy. If you disregard the binding tags, they are just regular go functions. They don't depend on the framework at all, so you can test them as you would any regular function.

Here's an example of testing a hypothetical `loginEndpoint` function:

```go
import "testing"

func TestLogin(t *testing.T) {

    t.Run("valid input returns token", func(t *testing.T) {
        output, err := loginEndpoint(loginInput{
            Email: "valid.email@example.org",
            Password: "some-valid-password",
        })
        if err != nil {
            t.Errorf("unexpected error: %s", err)
        }
        if output.Token == "" {
            t.Errorf("token was not returned")
        }
    })

    t.Run("invalid input returns error", func(t *testing.T) {
        _, err := loginEndpoint(loginInput{})
        if err == nil {
            t.Error("did not return an error")
        }
    })
}
```
