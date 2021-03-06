# Gap

Gap is a web framework aimed at removing HTTP boilerplate from your application.

Here's a minimal example of an application with one endpoint:

```go
package main

import (
    "github.com/hugollm/gap"
)

func main() {
    app := gap.New()
    app.Route("GET", "/", helloEndpoint)
    app.Run()
}
```

And here's the endpoint code:

```go
type helloInput struct {
    UserAgent string `header:"user-agent"`
}

type helloOutput struct {
    Message string `json:"message"`
}

func helloEndpoint(input helloInput) (helloOutput, error) {
    return helloOutput{"hello " + input.UserAgent}, nil
}
```

This endpoint gets the `User-Agent` header from the request and outputs a hello message on the response body as JSON.

```json
GET /
User-Agent: golang
---
200 {"message": "hello golang"}
```

Note how the endpoint is a "pure" function that **does not depend on the framework**.
Tag annotations bind inputs and outputs to requests and responses.
This makes it much easier to write and test your app!
