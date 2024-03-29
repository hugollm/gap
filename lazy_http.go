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
	parsedJSON  map[string]interface{}
	body        io.Reader
}

type requestError struct {
	Status  int    `response:"status"`
	Message string `response:"json,error"`
}

func (err requestError) Error() string {
	return err.Message
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

func (request *lazyRequest) getJSON(key string) interface{} {
	if request.parsedJSON == nil {
		body, err := ioutil.ReadAll(request.httpRequest.Body)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(body, &request.parsedJSON); err != nil {
			panic(requestError{400, "invalid json"})
		}
	}
	return request.parsedJSON[key]
}

type lazyResponse struct {
	httpResponse http.ResponseWriter
	jsonMap      map[string]interface{}
	status       int
	body         io.Reader
}

func newLazyResponse(httpResponse http.ResponseWriter) *lazyResponse {
	return &lazyResponse{httpResponse: httpResponse, status: 200}
}

func (response *lazyResponse) setJSON(key string, value interface{}) {
	if response.jsonMap == nil {
		response.jsonMap = map[string]interface{}{}
	}
	response.jsonMap[key] = value
}

func (response *lazyResponse) send() {
	response.httpResponse.WriteHeader(response.status)
	if response.body != nil {
		io.Copy(response.httpResponse, response.body)
		return
	}
	if response.jsonMap != nil {
		body, err := json.Marshal(response.jsonMap)
		if err != nil {
			panic(err)
		}
		response.httpResponse.Write(body)
	}
}
