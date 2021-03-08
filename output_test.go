package gap

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

type tErr struct {
	Status  int    `response:"status"`
	Message string `response:"json,message"`
}

func (err tErr) Error() string {
	return err.Message
}

func TestOutput(t *testing.T) {

	t.Run("can output to headers", func(t *testing.T) {
		type tIn struct{}
		type tOut struct {
			ContentType  string `response:"header,Content-Type"`
			CacheControl string `response:"header,Cache-Control"`
		}
		fn := func(input tIn) (tOut, error) {
			return tOut{"application/json", "no-cache"}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello", nil)
		response := httptest.NewRecorder()
		ep.handle(request, response)
		if response.Result().Header.Get("content-type") != "application/json" {
			t.Error("failed to set content-type header")
		}
		if response.Result().Header.Get("cache-control") != "no-cache" {
			t.Error("failed to set cache-control header")
		}
	})

	t.Run("can output to json", func(t *testing.T) {
		type tIn struct{}
		type tOut struct {
			Title  string `response:"json,title"`
			Public bool   `response:"json,public"`
		}
		fn := func(input tIn) (tOut, error) {
			return tOut{"lorem ipsum", true}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello", nil)
		response := httptest.NewRecorder()
		ep.handle(request, response)
		output := tOut{}
		json.Unmarshal(response.Body.Bytes(), &output)
		if response.Header().Get("content-type") != "application/json" {
			t.Error("failed to set content-type header")
		}
		if output.Title != "lorem ipsum" || output.Public != true {
			t.Error("failed to output json body")
		}
	})

	t.Run("can output to http status", func(t *testing.T) {
		type tIn struct{}
		type tOut struct {
			Status int `response:"status"`
		}
		fn := func(input tIn) (tOut, error) {
			return tOut{201}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello", nil)
		response := httptest.NewRecorder()
		ep.handle(request, response)
		if response.Code != 201 {
			t.Error("failed to output http status")
		}
	})

	t.Run("can output reader to body", func(t *testing.T) {
		type tIn struct{}
		type tOut struct {
			Body io.Reader `response:"body"`
		}
		fn := func(input tIn) (tOut, error) {
			body := strings.NewReader("lorem ipsum")
			return tOut{body}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello", nil)
		response := httptest.NewRecorder()
		ep.handle(request, response)
		output := tOut{}
		json.Unmarshal(response.Body.Bytes(), &output)
		if response.Body.String() != "lorem ipsum" {
			t.Error("failed to output reader to body")
		}
	})

	t.Run("outputs endpoint error as bad request", func(t *testing.T) {
		type tIn struct{}
		type tOut struct{}
		fn := func(input tIn) (tOut, error) {
			return tOut{}, errors.New("validation error")
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello", nil)
		response := httptest.NewRecorder()
		ep.handle(request, response)
		output := map[string]string{}
		json.Unmarshal(response.Body.Bytes(), &output)
		if response.Result().StatusCode != 400 {
			t.Error("failed to set status code")
		}
		if len(output) != 1 || output["error"] != "validation error" {
			t.Error("failed to output json body with error")
		}
	})

	t.Run("error can be an output struct", func(t *testing.T) {
		type tIn struct{}
		type tOut struct{}
		fn := func(input tIn) (tOut, error) {
			return tOut{}, tErr{401, "auth error"}
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello", nil)
		response := httptest.NewRecorder()
		ep.handle(request, response)
		output := map[string]string{}
		json.Unmarshal(response.Body.Bytes(), &output)
		if response.Result().StatusCode != 401 {
			t.Error("failed to set status code")
		}
		if len(output) != 1 || output["message"] != "auth error" {
			t.Error("failed to output json body with message")
		}
	})

	t.Run("empty spaces are ignored on tag values", func(t *testing.T) {
		type tIn struct{}
		type tOut struct {
			ContentType string `response:"header, content-type "`
		}
		fn := func(input tIn) (tOut, error) {
			return tOut{"image/jpeg"}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("POST", "/post", nil)
		request.Header.Set("content-type", "application/json")
		response := httptest.NewRecorder()
		ep.handle(request, response)
		if response.Header().Get("content-type") != "image/jpeg" {
			t.Error("failed to output content-type header")
		}
	})
}
