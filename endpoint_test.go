package gap

import (
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

func assertDoesNotPanic(t *testing.T) {
	err := recover()
	if err != nil {
		t.Errorf("did panic")
	}
}

func assertPanics(t *testing.T, msg string) {
	ierr := recover()
	if ierr == nil {
		t.Errorf("did not panic")
	} else {
		err, ok := ierr.(error)
		if ok {
			if err.Error() != msg {
				t.Errorf(`panic message was wrong: "%s"`, err.Error())
			}
		} else {
			panic(ierr)
		}
	}
}
