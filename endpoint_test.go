package gap

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestEndpoint(t *testing.T) {

	t.Run("can be constructed from function featuring any one of the accepted interfaces", func(t *testing.T) {
		defer assertDoesNotPanic(t)
		newEndpoint(func() {})
		newEndpoint(func() struct{} { return struct{}{} })
		newEndpoint(func() error { return nil })
		newEndpoint(func() (struct{}, error) { return struct{}{}, nil })
		newEndpoint(func(input struct{}) {})
		newEndpoint(func(input struct{}) struct{} { return struct{}{} })
		newEndpoint(func(input struct{}) error { return nil })
		newEndpoint(func(input struct{}) (struct{}, error) { return struct{}{}, nil })
	})

	t.Run("cannot be constructed from functions with invalid interfaces", func(t *testing.T) {
		functions := []interface{}{
			func(input struct{}, input2 struct{}) {},
			func() (struct{}, error, error) { return struct{}{}, nil, nil },
			func(input string) {},
			func() string { return "" },
			func() (struct{}, struct{}) { return struct{}{}, struct{}{} },
			func() (error, error) { return nil, nil },
		}
		for i, fn := range functions {
			t.Run(fmt.Sprintf("function %d", i+1), func(t *testing.T) {
				defer assertPanics(t, "invalid endpoint interface")
				newEndpoint(fn)
			})
		}
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

	t.Run("dummy endpoints yield 200 with no body", func(t *testing.T) {
		functions := []interface{}{
			func() {},
			func() struct{} { return struct{}{} },
			func() error { return nil },
			func() (struct{}, error) { return struct{}{}, nil },
			func(input struct{}) {},
			func(input struct{}) struct{} { return struct{}{} },
			func(input struct{}) error { return nil },
			func(input struct{}) (struct{}, error) { return struct{}{}, nil },
		}
		for i, fn := range functions {
			t.Run(fmt.Sprintf("function %d", i+1), func(t *testing.T) {
				ep := newEndpoint(fn)
				request := httptest.NewRequest("GET", "/", nil)
				response := httptest.NewRecorder()
				ep.handle(request, response)
				if response.Code != 200 {
					t.Error("failed to set status code")
				}
				if response.Body.String() != "" {
					t.Errorf("failed to send an empty body: %s", response.Body.String())
				}
			})
		}
	})

	t.Run("endpoints with different interfaces yield expected responses", func(t *testing.T) {
		type tIn struct {
			UserAgent string `request:"header,user-agent"`
		}
		type tOut struct {
			Hello string `response:"json,hello"`
		}
		type testCase struct {
			name     string
			function interface{}
			status   int
			body     string
		}
		cases := []testCase{
			testCase{"no input or output", func() {}, 200, ""},
			testCase{"output only", func() tOut { return tOut{"world"} }, 200, `{"hello":"world"}`},
			testCase{"error only", func() error { return errors.New("ops") }, 400, `{"error":"ops"}`},
			testCase{"output and error, returning output", func() (tOut, error) { return tOut{"world"}, nil }, 200, `{"hello":"world"}`},
			testCase{"output and error, returning error", func() (tOut, error) { return tOut{}, errors.New("ops") }, 400, `{"error":"ops"}`},
			testCase{"input only", func(input tIn) {}, 200, ""},
			testCase{"input and output", func(input tIn) tOut { return tOut{input.UserAgent} }, 200, `{"hello":"test"}`},
			testCase{"input and error, returning error", func(input tIn) error { return errors.New(input.UserAgent) }, 400, `{"error":"test"}`},
			testCase{"input and error, returning nil", func(input tIn) error { return nil }, 200, ""},
			testCase{"input, output and error, returning output", func(input tIn) (tOut, error) { return tOut{"world"}, nil }, 200, `{"hello":"world"}`},
			testCase{"input, output and error, returning error", func(input tIn) (tOut, error) { return tOut{}, errors.New(input.UserAgent) }, 400, `{"error":"test"}`},
		}
		for _, tcase := range cases {
			t.Run(fmt.Sprintf(tcase.name), func(t *testing.T) {
				ep := newEndpoint(tcase.function)
				request := httptest.NewRequest("GET", "/", nil)
				request.Header.Set("user-agent", "test")
				response := httptest.NewRecorder()
				ep.handle(request, response)
				if response.Code != tcase.status {
					t.Errorf("unexpected status code: %d != %d", response.Code, tcase.status)
				}
				if response.Body.String() != tcase.body {
					t.Errorf("unexpected response body: %s != %s", response.Body.String(), tcase.body)
				}
			})
		}
	})
}
