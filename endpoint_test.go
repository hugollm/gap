package gap

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEndpoint(t *testing.T) {

	t.Run("can be constructed from func with right interface", func(t *testing.T) {
		defer assertDoesNotPanic(t)
		newEndpoint(hello)
	})

	t.Run("cannot be constructed from func with wrong interface", func(t *testing.T) {

		t.Run("wrong number of arguments", func(t *testing.T) {
			defer assertPanics(t, "endpoint interface must be: func(struct) (struct, error)")
			newEndpoint(func(input helloInput) helloOutput { return helloOutput{} })
		})

		t.Run("input is not struct", func(t *testing.T) {
			defer assertPanics(t, "endpoint interface must be: func(struct) (struct, error)")
			newEndpoint(func(input string) (helloOutput, error) { return helloOutput{}, nil })
		})

		t.Run("first output is not struct", func(t *testing.T) {
			defer assertPanics(t, "endpoint interface must be: func(struct) (struct, error)")
			newEndpoint(func(input helloInput) (string, error) { return "hello", nil })
		})

		t.Run("second output is not error", func(t *testing.T) {
			defer assertPanics(t, "endpoint interface must be: func(struct) (struct, error)")
			newEndpoint(func(input helloInput) (helloOutput, string) { return helloOutput{}, "error" })
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
	})
}

type helloInput struct {
	Name string `json:"name"`
}

type helloOutput struct {
	Message string `json:"message"`
}

func hello(input helloInput) (helloOutput, error) {
	return helloOutput{Message: "hello " + input.Name}, nil
}
