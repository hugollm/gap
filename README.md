# Gap

Gap is a web framework aimed at removing HTTP boilerplate from your application.


## Overview

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

type helloInput struct {
    UserAgent string `request:"header,user-agent"`
}

type helloOutput struct {
    Message string `response:"json,message"`
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

Note how the endpoint function **does not depend on the framework**. Tag annotations take care of binding inputs and outputs to requests and responses. This makes it much easier to write and test your app.
