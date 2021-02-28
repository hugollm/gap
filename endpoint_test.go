package gap

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEndpoint(t *testing.T) {

	t.Run("can be constructed from func with right interface", func(t *testing.T) {
		defer assertDoesNotPanic(t)
		type tIn struct{}
		type tOut struct{}
		newEndpoint(func(input tIn) (tOut, error) { return tOut{}, nil })
	})

	t.Run("cannot be constructed from func with wrong interface", func(t *testing.T) {

		type tIn struct{}
		type tOut struct{}

		t.Run("wrong number of arguments", func(t *testing.T) {
			defer assertPanics(t, "endpoint interface must be: func(struct) (struct, error)")
			newEndpoint(func(input tIn) tOut { return tOut{} })
		})

		t.Run("input is not struct", func(t *testing.T) {
			defer assertPanics(t, "endpoint interface must be: func(struct) (struct, error)")
			newEndpoint(func(input string) (tOut, error) { return tOut{}, nil })
		})

		t.Run("first output is not struct", func(t *testing.T) {
			defer assertPanics(t, "endpoint interface must be: func(struct) (struct, error)")
			newEndpoint(func(input tIn) (string, error) { return "hello", nil })
		})

		t.Run("second output is not error", func(t *testing.T) {
			defer assertPanics(t, "endpoint interface must be: func(struct) (struct, error)")
			newEndpoint(func(input tIn) (tOut, string) { return tOut{}, "error" })
		})

		t.Run("missing bind on input field", func(t *testing.T) {
			defer assertPanics(t, "missing bind on input field")
			type tIn struct{ Name string }
			newEndpoint(func(input tIn) (tOut, error) { return tOut{}, nil })
		})

		t.Run("missing bind on output field", func(t *testing.T) {
			defer assertPanics(t, "missing bind on output field")
			type tOut struct{ Name string }
			newEndpoint(func(input tIn) (tOut, error) { return tOut{}, nil })
		})
	})

	t.Run("handle calls underlying function", func(t *testing.T) {
		called := false
		type tIn struct{}
		type tOut struct{}
		fn := func(input tIn) (tOut, error) {
			called = true
			return tOut{}, nil
		}
		ep := newEndpoint(fn)
		request := httptest.NewRequest("GET", "/hello", nil)
		response := httptest.NewRecorder()
		ep.handle(request, response)
		if !called {
			t.Error("endpoint was not called")
		}
	})

	t.Run("can handle request input", func(t *testing.T) {

		t.Run("can get input from headers", func(t *testing.T) {
			type tIn struct {
				Auth        string `header:"authorization"`
				ContentType string `header:"content-type"`
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

		t.Run("can get input from query string", func(t *testing.T) {
			type tIn struct {
				Limit string `query:"limit"`
				Page  string `query:"page"`
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
				Title  string `json:"title"`
				Public bool   `json:"public"`
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
				HeaderAuth string `header:"auth"`
				QueryAuth  string `query:"auth"`
				JsonAuth   string `json:"auth"`
			}
			type tOut struct{}
			fn := func(input tIn) (tOut, error) {
				if input.HeaderAuth != "hauth" || input.QueryAuth != "qauth" || input.JsonAuth != "jauth" {
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
				Body io.Reader `body:"*"`
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
	})

	t.Run("can handle request output", func(t *testing.T) {

		t.Run("can output to headers", func(t *testing.T) {
			type tIn struct{}
			type tOut struct {
				ContentType  string `header:"Content-Type"`
				CacheControl string `header:"Cache-Control"`
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
				Title  string `json:"title"`
				Public bool   `json:"public"`
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
			if output.Title != "lorem ipsum" || output.Public != true {
				t.Error("failed to output json body")
			}
		})

		t.Run("can output to http status", func(t *testing.T) {
			type tIn struct{}
			type tOut struct {
				Status int `status:"*"`
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
				Body io.Reader `body:"*"`
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
	})
}

type tErr struct {
	Status  int    `status:"*"`
	Message string `json:"message"`
}

func (err tErr) Error() string {
	return err.Message
}
