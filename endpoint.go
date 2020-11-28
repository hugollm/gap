package gap

import (
	"errors"
	"net/http"
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

func (ep *endpoint) handle(request *http.Request, response http.ResponseWriter) {
	input := reflect.New(ep.rtype.In(0)).Elem()
	for i := 0; i < ep.rtype.In(0).NumField(); i++ {
		field := ep.rtype.In(0).Field(i)
		if header, ok := field.Tag.Lookup("header"); ok {
			input.FieldByName(field.Name).SetString(request.Header.Get(header))
		}
		if query, ok := field.Tag.Lookup("query"); ok {
			input.FieldByName(field.Name).SetString(request.URL.Query().Get(query))
		}
	}
	ep.rval.Call([]reflect.Value{input})
}
