# Output

Tags are used to bind your output fields to responses. This article describes all the response bindings available.

Here's a list of all the output tag formats:

```
Header  response:"header,name"
JSON    response:"json,name"
Status  response:"status"
Body    response:"body"
```

## Header

Used to send a header to response. Example:

```go
type struct output {
    ContentType string `response:"header,Content-Type"`
}
```

Headers are case insensitive, so `content-type` would also work.


## JSON

Used to send JSON values on the response body.

```go
type struct output {
    Page int `response:"json,page"`
}
```

Like on inputs, types are flexible on the json binding. You can use any type that would be normally serializable with JSON. Note how your output will always be inside an object `{...}`. This is a limitation, but comes with some benefits. Always sending objects means that your output fields will have names, and adding new fields is possible without breaking contract with the API clients.

Inside the object, any JSON structure is valid. Make sure you properly use `json:"..."` tags on the nested structures.


## Status

Used to send a response with different status code.

```go
type struct output {
    Status int `response:"status"`
}
```

Normally, an endpoint will response with the following status codes:

* 200: when there's no error
* 400: when you return an error from the endpoint
* 500: in case of panics

So this binding is basically about overriding that with an arbitrary code.


## Body

Used to send an arbitrary body stream with the response.

```go
import "io"

type struct output {
    Body io.Reader `response:"body"`
}
```

The body is an `io.Reader` so you don't need to put all the bytes in memory at once. File downloads are a common use case.
