# Endpoints

Endpoints expose functionality on your API. You add endpoints to your `App` like this:

```go
app.Route("GET", "/hello", helloEndpoint)
```

In the example above, `helloEndpoint` is just a function that impelents one of the valid interfaces:

```go
func()
func() struct
func() error
func() (struct, error)
func(input struct)
func(input struct) struct
func(input struct) error
func(input struct) (struct, error)
```

As you can see from the interfaces, the endpoints can accept/return:

* Optional input struct
* Optional output struct
* Optional error

The simplest endpoint you can write is one that have no inputs or outputs:

```go
func endpoint() {
    println("some side effect...")
}
```

And here's an example that's closer to what you'll see in real life:

```go
type loginInput struct {
    Email string `request:"json,email"`
    Password string `request:"json,password"`
}

type loginOutput struct {
    Token string `response:"json,token"`
}

func loginEndpoint(input loginInput) (loginOutput, error) {
    token, err := validateEmailAndPassword(input.Email, input.Password)
    if err != nil {
        return loginOutput{}, err
    }
    return loginOutput{token}, nil
}
```

In the hypothetical example above, the endpoint two responses. If there's no error, it will send:

```
200 OK
{"token": "some-session-token"}
```

If an error is returned, response will be:

```
400 Bad Request
{"error": "message from err.Error"}
```

Inputs and outputs are further detailed next on this guide.
