package gap

import (
	"testing"
)

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
