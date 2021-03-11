# App

The `App` struct is the main glue that holds your application together. It contains your endpoints and runs your API. Here's a minimal example of the simplest app you can run:

```go
package main

import (
    "github.com/hugollm/gap"
)

func main() {
    app := gap.New()
    app.Run()
}
```

The method `Run` is a shorcut that listens on `localhost:8000`. Since there's no endpoints on this app, it will just answer 404 for all requests. We cover endpoints next on this guide.

When running on production you might want to have more control over the server. `App` implements the [http.Handler](https://golang.org/pkg/net/http/#Handler) interface. This means you can seamlessly use it with go's native web server:

```go
package main

import (
    "net/http"
    "github.com/hugollm/gap"
)

func main() {
    app := gap.New()
    http.ListenAndServe(":8000", app)
}
```

Or with even more configuration:

```go
package main

import (
    "net/http"
    "time"
    "github.com/hugollm/gap"
)

func main() {
    app := gap.New()
    server := &http.Server{
        Addr:           ":8000",
        Handler:        app,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 10 * 1024,
    }
    server.ListenAndServe()
}
```
