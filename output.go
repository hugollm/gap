package gap

import (
	"io"
	"reflect"
)

type outputField interface {
	write(response *lazyResponse, value reflect.Value)
}

func newOutputField(field reflect.StructField) outputField {
	if key, ok := field.Tag.Lookup("header"); ok {
		return headerOutput{key}
	}
	if key, ok := field.Tag.Lookup("json"); ok {
		return jsonOutput{key}
	}
	if _, ok := field.Tag.Lookup("body"); ok {
		return bodyOutput{}
	}
	panic("invalid output field")
}

type headerOutput struct {
	key string
}

func (output headerOutput) write(response *lazyResponse, value reflect.Value) {
	str, ok := value.Interface().(string)
	if ok {
		response.httpResponse.Header().Add(output.key, str)
	}
}

type jsonOutput struct {
	key string
}

func (output jsonOutput) write(response *lazyResponse, value reflect.Value) {
	response.setJson(output.key, value.Interface())
}

type bodyOutput struct{}

func (output bodyOutput) write(response *lazyResponse, value reflect.Value) {
	response.body = value.Interface().(io.Reader)
}
