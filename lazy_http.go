package gap

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type lazyRequest struct {
	httpRequest *http.Request
	parsedQuery url.Values
	parsedJson  map[string]interface{}
	body        io.Reader
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
	status       int
	body         io.Reader
}

func newLazyResponse(httpResponse http.ResponseWriter) *lazyResponse {
	return &lazyResponse{httpResponse: httpResponse, jsonMap: map[string]interface{}{}, status: 200}
}

func (response *lazyResponse) setJson(key string, value interface{}) {
	response.jsonMap[key] = value
}

func (response *lazyResponse) send() {
	response.httpResponse.WriteHeader(response.status)
	if response.body != nil {
		io.Copy(response.httpResponse, response.body)
		return
	}
	body, err := json.Marshal(response.jsonMap)
	if err != nil {
		panic(err)
	}
	response.httpResponse.Write(body)
}
