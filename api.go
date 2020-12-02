package gap

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type api struct {
	routes map[string]route
}

type route struct {
	method   string
	endpoint endpoint
}

type errorResponse struct {
	status int
	body   interface{}
}

func New() *api {
	return &api{routes: map[string]route{}}
}

func Response(status int, body interface{}) errorResponse {
	return errorResponse{status, body}
}

func (api *api) Route(method string, path string, fn interface{}) {
	api.routes[path] = route{method, newEndpoint(fn)}
}

func (api *api) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	defer writeErrorOnPanic(response)
	route, found := api.routes[request.URL.Path]
	if !found {
		writeNotFound(response)
		return
	}
	if route.method != request.Method {
		writeMethodNotAllowed(response)
		return
	}
	route.endpoint.handle(request, response)
}

func writeNotFound(response http.ResponseWriter) {
	response.WriteHeader(404)
	response.Write([]byte(`{"error":"not found"}`))
}

func writeMethodNotAllowed(response http.ResponseWriter) {
	response.WriteHeader(405)
	response.Write([]byte(`{"error":"method not allowed"}`))
}

func writeErrorOnPanic(httpResponse http.ResponseWriter) {
	ierr := recover()
	if ierr != nil {
		if resp, ok := ierr.(errorResponse); ok {
			body, err := json.Marshal(resp.body)
			if err != nil {
				panic(err)
			}
			httpResponse.WriteHeader(resp.status)
			httpResponse.Write(body)
		} else {
			fmt.Printf("PANIC: %s", ierr)
			httpResponse.WriteHeader(500)
			httpResponse.Write([]byte(`{"error":"server error"}`))
		}
	}
}
