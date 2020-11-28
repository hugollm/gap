package gap

import (
	"errors"
	"reflect"
)

type endpoint struct {
	rval  reflect.Value
	rtype reflect.Type
}

func newEndpoint(function interface{}) endpoint {
	ep := endpoint{}
	ep.rval = reflect.ValueOf(function)
	ep.rtype = reflect.TypeOf(function)
	validateInterface(ep.rtype)
	return ep
}

func validateInterface(rtype reflect.Type) {
	if rtype.NumIn() != 1 ||
		rtype.NumOut() != 2 ||
		rtype.In(0).Kind() != reflect.Struct ||
		rtype.Out(0).Kind() != reflect.Struct ||
		!rtype.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		panic(errors.New("endpoint interface must be: func(struct) (struct, error)"))
	}
}
