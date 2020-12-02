package gap

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type lazyRequest struct {
	httpRequest *http.Request
	parsedQuery url.Values
	parsedJson  map[string]interface{}
}

func newLazyRequest(httpRequest *http.Request) *lazyRequest {
	return &lazyRequest{httpRequest: httpRequest}
}

func (request *lazyRequest) getQuery(key string) string {
	if request.parsedQuery == nil {
		request.parsedQuery = request.httpRequest.URL.Query()
	}
	return request.parsedQuery.Get(key)
}

func (request *lazyRequest) getJson(key string) interface{} {
	if request.parsedJson == nil {
		body, err := ioutil.ReadAll(request.httpRequest.Body)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(body, &request.parsedJson); err != nil {
			panic(Response(400, map[string]string{"error": "invalid json"}))
		}
	}
	return request.parsedJson[key]
}

type lazyResponse struct {
	httpResponse http.ResponseWriter
	jsonMap      map[string]interface{}
}

func newLazyResponse(httpResponse http.ResponseWriter) *lazyResponse {
	return &lazyResponse{httpResponse: httpResponse, jsonMap: map[string]interface{}{}}
}

func (response *lazyResponse) setJson(key string, value interface{}) {
	response.jsonMap[key] = value
}

func (response *lazyResponse) send(status int) {
	body, err := json.Marshal(response.jsonMap)
	if err != nil {
		panic(err)
	}
	response.httpResponse.WriteHeader(status)
	response.httpResponse.Write(body)
}
