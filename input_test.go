package gap

import (
	"io"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInput(t *testing.T) {

	t.Run("can get input from headers", func(t *testing.T) {
		type tIn struct {
			Auth        string `request:"header,authorization"`
			ContentType string `request:"header,content-type"`
		}
		type tOut struct{}
		fn := func(input tIn) (tOut, error) {
			if input.Auth != "token" || input.ContentType != "application/json" {
				t.Error("failed to fetch input headers")
			}
			return tOut{}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello", nil)
		request.Header.Set("Authorization", "token")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		ep.handle(request, response)
	})

	t.Run("can get input from whole path", func(t *testing.T) {
		type tIn struct {
			Path string `request:"path"`
		}
		fn := func(input tIn) {
			if input.Path != "/hello/world" {
				t.Errorf("failed to fetch path as input: %s", input.Path)
			}
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello/world?q=query", nil)
		response := httptest.NewRecorder()
		ep.handle(request, response)
	})

	t.Run("can get input from query string", func(t *testing.T) {
		type tIn struct {
			Limit string `request:"query,limit"`
			Page  string `request:"query,page"`
		}
		type tOut struct{}
		fn := func(input tIn) (tOut, error) {
			if input.Limit != "10" || input.Page != "2" {
				t.Error("failed to fetch input queries")
			}
			return tOut{}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello?limit=10&page=2", nil)
		response := httptest.NewRecorder()
		ep.handle(request, response)
	})

	t.Run("can get input from json body", func(t *testing.T) {
		type tIn struct {
			Title  string `request:"json,title"`
			Public bool   `request:"json,public"`
		}
		type tOut struct{}
		fn := func(input tIn) (tOut, error) {
			if input.Title != "lorem ipsum" || !input.Public {
				t.Error("failed to fetch input from json")
			}
			return tOut{}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello", strings.NewReader(`{"title": "lorem ipsum", "public": true}`))
		response := httptest.NewRecorder()
		ep.handle(request, response)
	})

	t.Run("can get input from multiple sources with the same name", func(t *testing.T) {
		type tIn struct {
			HeaderAuth string `request:"header,auth"`
			QueryAuth  string `request:"query,auth"`
			JSONAuth   string `request:"json,auth"`
		}
		type tOut struct{}
		fn := func(input tIn) (tOut, error) {
			if input.HeaderAuth != "hauth" || input.QueryAuth != "qauth" || input.JSONAuth != "jauth" {
				t.Error("failed to fetch input from multiple sources")
			}
			return tOut{}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello?auth=qauth", strings.NewReader(`{"auth": "jauth"}`))
		request.Header.Set("auth", "hauth")
		response := httptest.NewRecorder()
		ep.handle(request, response)
	})

	t.Run("can get whole request body as input", func(t *testing.T) {
		type tIn struct {
			Body io.Reader `request:"body"`
		}
		type tOut struct{}
		fn := func(input tIn) (tOut, error) {
			body, _ := ioutil.ReadAll(input.Body)
			if string(body) != "lorem ipsum" {
				t.Error("failed to fetch body as input")
			}
			return tOut{}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello?auth=qauth", strings.NewReader("lorem ipsum"))
		response := httptest.NewRecorder()
		ep.handle(request, response)
	})

	t.Run("empty spaces are ignored on tag values", func(t *testing.T) {
		type tIn struct {
			ContentType string `request:"header, content-type "`
		}
		type tOut struct{}
		fn := func(input tIn) (tOut, error) {
			if input.ContentType != "application/json" {
				t.Error("failed to input content-type header")
			}
			return tOut{}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("POST", "/post", nil)
		request.Header.Set("content-type", "application/json")
		response := httptest.NewRecorder()
		ep.handle(request, response)
	})
}
