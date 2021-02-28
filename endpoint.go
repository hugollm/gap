package gap

import (
	"errors"
	"net/http"
	"reflect"
)

type endpoint struct {
	rval      reflect.Value
	rtype     reflect.Type
	inFields  map[string]inputField
	outFields map[string]outputField
}

func newEndpoint(function interface{}) endpoint {
	ep := endpoint{}
	ep.rval = reflect.ValueOf(function)
	ep.rtype = reflect.TypeOf(function)
	validateInterface(ep.rtype)
	ep.setupInputFields()
	ep.setupOutputFields()
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

func (ep *endpoint) setupInputFields() {
	ep.inFields = map[string]inputField{}
	for i := 0; i < ep.rtype.In(0).NumField(); i++ {
		field := ep.rtype.In(0).Field(i)
		ep.inFields[field.Name] = newInputField(field)
	}
}

func (ep *endpoint) setupOutputFields() {
	ep.outFields = map[string]outputField{}
	for i := 0; i < ep.rtype.Out(0).NumField(); i++ {
		field := ep.rtype.Out(0).Field(i)
		ep.outFields[field.Name] = newOutputField(field)
	}
}

func (ep *endpoint) handle(request *http.Request, response http.ResponseWriter) {
	input := ep.readInput(request)
	result := ep.rval.Call([]reflect.Value{input})
	ep.writeOutput(response, result)
}

func (ep *endpoint) readInput(httpRequest *http.Request) reflect.Value {
	request := newLazyRequest(httpRequest)
	input := reflect.New(ep.rtype.In(0)).Elem()
	for name, field := range ep.inFields {
		target := input.FieldByName(name)
		target.Set(field.read(request).Convert(target.Type()))
	}
	return input
}

func (ep *endpoint) writeOutput(httpResponse http.ResponseWriter, result []reflect.Value) {
	response := newLazyResponse(httpResponse)
	output, outErr := result[0], result[1]
	if outErr.IsZero() {
		for name, field := range ep.outFields {
			field.write(response, output.FieldByName(name))
		}
		response.send()
	} else {
		response.status = 400
		err := outErr.Interface().(error)
		outErr = reflect.ValueOf(err)
		errFields := getErrorFields(err)
		if errFields != nil {
			for name, field := range errFields {
				field.write(response, outErr.FieldByName(name))
			}
		} else {
			response.setJson("error", outErr.Interface().(error).Error())
		}
		response.send()
	}
}

func getErrorFields(err error) map[string]outputField {
	rtype := reflect.TypeOf(err)
	if rtype.Kind() != reflect.Struct {
		return nil
	}
	errFields := map[string]outputField{}
	for i := 0; i < rtype.NumField(); i++ {
		field := rtype.Field(i)
		errFields[field.Name] = newOutputField(field)
	}
	return errFields
}
