package gap

import (
	"reflect"
)

type inputField interface {
	read(request *lazyRequest) reflect.Value
}

func newInputField(field reflect.StructField) inputField {
	if key, ok := field.Tag.Lookup("header"); ok {
		return headerInput{key}
	}
	if key, ok := field.Tag.Lookup("query"); ok {
		return queryInput{key}
	}
	if key, ok := field.Tag.Lookup("json"); ok {
		return jsonInput{key}
	}
	if _, ok := field.Tag.Lookup("body"); ok {
		return bodyInput{}
	}
	panic("invalid input field")
}

type headerInput struct {
	key string
}

func (input headerInput) read(request *lazyRequest) reflect.Value {
	return reflect.ValueOf(request.httpRequest.Header.Get(input.key))
}

type queryInput struct {
	key string
}

func (input queryInput) read(request *lazyRequest) reflect.Value {
	return reflect.ValueOf(request.getQuery(input.key))
}

type jsonInput struct {
	key string
}

func (input jsonInput) read(request *lazyRequest) reflect.Value {
	return reflect.ValueOf(request.getJson(input.key))
}

type bodyInput struct{}

func (input bodyInput) read(request *lazyRequest) reflect.Value {
	return reflect.ValueOf(request.httpRequest.Body)
}
