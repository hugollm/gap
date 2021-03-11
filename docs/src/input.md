# Input

Tags are used to bind request data to your input structs so your endpoint doesn't have to deal with it. This article describes all the request bindings available.

Here's a list of all the tag formats:

```
Header  request:"header,name"
Path    request:"path"
Query   request:"query,name"
JSON    request:"json,name"
Body    request:"body"
```

## Header

Used to retrieve headers form the request. Example:

```go
type struct input {
    ContentType string `request:"header,Content-Type"`
}
```

Headers are case insensitive, so `content-type` would also work.


## Path

Used to retrieve the whole path from the request URL. Example:

```go
type struct input {
    Path string `request:"path"`
}
```

Path will then contain something like `/path/to/endpoint`.


## Query

Used to retrieve params from que URL query string. Example:

```go
type struct input {
    Page string `request:"query,page"`
}
```

Note how the type is `string` even though it probably contains a number. At this point, the framework does not perform implicit conversions automatically.


## JSON

Used to retrieve values from a JSON body.

```go
type struct input {
    Page int `request:"json,page"`
}
```

In this case, the types are as flexible as a regular JSON parse can be. Note however that only json objects will work (e.g. `{...}`). Inside the object, any valid JSON structure is allowed (even nested). Beyond the first nesting level, you'll need to properly make use of the `json:""` tag, like you would normally outside of the framework.


## Body

Used to bind the whole request body stream.

```go
import "io"

type struct input {
    Body io.Reader `request:"body"`
}
```

The body is retrieved as an `io.Reader` so you don't need to put all the bytes in memory at once. File uploads are a common use case.
