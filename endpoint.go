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
	validateEndpointInterface(ep.rtype)
	ep.setupInputFields()
	ep.setupOutputFields()
	return ep
}

func validateEndpointInterface(rtype reflect.Type) {
	if rtype.NumIn() > 1 ||
		rtype.NumOut() > 2 ||
		(rtype.NumIn() == 1 && !typeIsStruct(rtype.In(0))) ||
		(rtype.NumOut() == 1 && (!typeIsStruct(rtype.Out(0)) && !typeIsError(rtype.Out(0)))) ||
		(rtype.NumOut() == 2 && (!typeIsStruct(rtype.Out(0)) || !typeIsError(rtype.Out(1)))) {
		panic(errors.New("invalid endpoint interface"))
	}
}

func typeIsStruct(rtype reflect.Type) bool {
	return rtype.Kind() == reflect.Struct
}

func typeIsError(rtype reflect.Type) bool {
	return rtype.Implements(reflect.TypeOf((*error)(nil)).Elem())
}

func (ep *endpoint) setupInputFields() {
	if ep.rtype.NumIn() == 0 {
		return
	}
	ep.inFields = map[string]inputField{}
	for i := 0; i < ep.rtype.In(0).NumField(); i++ {
		field := ep.rtype.In(0).Field(i)
		ep.inFields[field.Name] = newInputField(field)
	}
}

func (ep *endpoint) setupOutputFields() {
	if ep.rtype.NumOut() == 0 || ep.rtype.Out(0).Kind() != reflect.Struct {
		return
	}
	ep.outFields = map[string]outputField{}
	for i := 0; i < ep.rtype.Out(0).NumField(); i++ {
		field := ep.rtype.Out(0).Field(i)
		ep.outFields[field.Name] = newOutputField(field)
	}
}

func (ep *endpoint) handle(request *http.Request, httpResponse http.ResponseWriter) {
	defer ep.writeErrorOnPanic(httpResponse)
	input := ep.readInput(request)
	result := ep.rval.Call(input)
	ep.writeResponse(httpResponse, result)
}

func (ep *endpoint) readInput(httpRequest *http.Request) []reflect.Value {
	if ep.rtype.NumIn() == 0 {
		return nil
	}
	request := newLazyRequest(httpRequest)
	input := reflect.New(ep.rtype.In(0)).Elem()
	for name, field := range ep.inFields {
		target := input.FieldByName(name)
		target.Set(field.read(request).Convert(target.Type()))
	}
	return []reflect.Value{input}
}

func (ep *endpoint) writeResponse(httpResponse http.ResponseWriter, result []reflect.Value) {
	if ep.rtype.NumOut() == 0 {
		return
	} else if ep.rtype.NumOut() == 1 {
		if typeIsStruct(ep.rtype.Out(0)) {
			rvOut := result[0]
			ep.writeOutput(httpResponse, rvOut)
		} else if typeIsError(ep.rtype.Out(0)) {
			rvErr := result[0]
			if !rvErr.IsNil() {
				ep.writeError(httpResponse, rvErr.Elem())
			}
		}
	} else if ep.rtype.NumOut() == 2 {
		rvOut, rvErr := result[0], result[1]
		if rvErr.IsNil() {
			ep.writeOutput(httpResponse, rvOut)
		} else {
			ep.writeError(httpResponse, rvErr.Elem())
		}
	}
}

func (ep *endpoint) writeOutput(httpResponse http.ResponseWriter, rvOut reflect.Value) {
	if len(ep.outFields) == 0 {
		return
	}
	response := newLazyResponse(httpResponse)
	for name, field := range ep.outFields {
		field.write(response, rvOut.FieldByName(name))
	}
	response.send()
}

func (ep *endpoint) writeError(httpResponse http.ResponseWriter, rvErr reflect.Value) {
	response := newLazyResponse(httpResponse)
	response.status = 400
	errFields := getErrorFields(rvErr)
	if errFields != nil {
		for name, field := range errFields {
			field.write(response, rvErr.FieldByName(name))
		}
	} else {
		response.setJSON("error", rvErr.Interface().(error).Error())
	}
	response.send()
}

func getErrorFields(rvErr reflect.Value) map[string]outputField {
	rtErr := rvErr.Type()
	if rtErr.Kind() != reflect.Struct {
		return nil
	}
	errFields := map[string]outputField{}
	for i := 0; i < rtErr.NumField(); i++ {
		field := rtErr.Field(i)
		errFields[field.Name] = newOutputField(field)
	}
	return errFields
}

func (ep *endpoint) writeErrorOnPanic(httpResponse http.ResponseWriter) {
	ierr := recover()
	if ierr != nil {
		rvErr := reflect.ValueOf(ierr)
		if isOutputStruct(rvErr) {
			ep.writeError(httpResponse, rvErr)
		} else {
			panic(ierr)
		}
	}
}

func isOutputStruct(rvStruct reflect.Value) bool {
	rtStruct := rvStruct.Type()
	if rtStruct.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < rtStruct.NumField(); i++ {
		field := rtStruct.Field(i)
		if field.Tag.Get("response") != "" {
			return true
		}
	}
	return false
}
