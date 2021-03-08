package gap

import (
	"errors"
	"io"
	"reflect"
)

type outputField interface {
	write(response *lazyResponse, value reflect.Value)
}

func newOutputField(field reflect.StructField) outputField {
	tagParts := splitTag(field.Tag.Get("response"))
	if len(tagParts) == 2 && tagParts[0] == "header" {
		return headerOutput{tagParts[1]}
	}
	if len(tagParts) == 2 && tagParts[0] == "json" {
		return jsonOutput{tagParts[1]}
	}
	if len(tagParts) == 1 && tagParts[0] == "status" {
		return statusOutput{}
	}
	if len(tagParts) == 1 && tagParts[0] == "body" {
		return bodyOutput{}
	}
	panic(errors.New("missing or invalid response tag on output field"))
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
	response.httpResponse.Header().Set("Content-Type", "application/json")
	response.setJson(output.key, value.Interface())
}

type statusOutput struct{}

func (output statusOutput) write(response *lazyResponse, value reflect.Value) {
	response.status = value.Interface().(int)
}

type bodyOutput struct{}

func (output bodyOutput) write(response *lazyResponse, value reflect.Value) {
	response.body = value.Interface().(io.Reader)
}
