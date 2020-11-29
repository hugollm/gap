package gap

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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
	defer writeServerErrorOnPanic(response)
	input := ep.readInput(request)
	result := ep.rval.Call([]reflect.Value{input.Elem()})
	ep.writeOutput(response, result)
}

func writeServerErrorOnPanic(response http.ResponseWriter) {
	err := recover()
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(`{"error":"server error"}`))
	}
}

func (ep *endpoint) readInput(request *http.Request) reflect.Value {
	input := reflect.New(ep.rtype.In(0))
	jsonWasRead := false
	for i := 0; i < ep.rtype.In(0).NumField(); i++ {
		field := ep.rtype.In(0).Field(i)
		if header, ok := field.Tag.Lookup("header"); ok {
			input.Elem().FieldByName(field.Name).SetString(request.Header.Get(header))
		}
		if query, ok := field.Tag.Lookup("query"); ok {
			input.Elem().FieldByName(field.Name).SetString(request.URL.Query().Get(query))
		}
		if _, ok := field.Tag.Lookup("json"); ok && !jsonWasRead {
			body, err := ioutil.ReadAll(request.Body)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(body, input.Interface()); err != nil {
				panic(err)
			}
			jsonWasRead = true
		}
	}
	return input
}

func (ep *endpoint) writeOutput(response http.ResponseWriter, result []reflect.Value) {
	output, outErr := result[0], result[1]
	bodyMap := map[string]interface{}{}
	if outErr.Interface() == nil {
		for i := 0; i < ep.rtype.Out(0).NumField(); i++ {
			field := ep.rtype.Out(0).Field(i)
			if header, ok := field.Tag.Lookup("header"); ok {
				if hval, ok := output.FieldByName(field.Name).Interface().(string); ok {
					response.Header().Add(header, hval)
				}
			}
			if jtag, ok := field.Tag.Lookup("json"); ok {
				bodyMap[jtag] = output.FieldByName(field.Name).Interface()
			}
		}
	} else {
		response.WriteHeader(400)
		bodyMap["error"] = outErr.Interface().(error).Error()
	}
	jsonBody, err := json.Marshal(bodyMap)
	if err != nil {
		panic(err)
	}
	response.Write(jsonBody)
}
