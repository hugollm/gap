package gap

import (
	"net/http/httptest"
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
			defer assertPanics(t, "missing or invalid request tag on input field")
			type tIn struct{ Name string }
			newEndpoint(func(input tIn) (tOut, error) { return tOut{}, nil })
		})

		t.Run("missing bind on output field", func(t *testing.T) {
			defer assertPanics(t, "missing or invalid response tag on output field")
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
}
