package gap

import (
	"errors"
	"reflect"
	"strings"
)

type inputField interface {
	read(request *lazyRequest) reflect.Value
}

func newInputField(field reflect.StructField) inputField {
	tagParts := splitTag(field.Tag.Get("request"))
	if len(tagParts) == 2 && tagParts[0] == "header" {
		return headerInput{tagParts[1]}
	}
	if len(tagParts) == 1 && tagParts[0] == "path" {
		return pathInput{}
	}
	if len(tagParts) == 2 && tagParts[0] == "query" {
		return queryInput{tagParts[1]}
	}
	if len(tagParts) == 2 && tagParts[0] == "json" {
		return jsonInput{tagParts[1]}
	}
	if len(tagParts) == 1 && tagParts[0] == "body" {
		return bodyInput{}
	}
	panic(errors.New("missing or invalid request tag on input field"))
}

func splitTag(tag string) []string {
	tagParts := strings.Split(tag, ",")
	for i, part := range tagParts {
		tagParts[i] = strings.TrimSpace(part)
	}
	return tagParts
}

type headerInput struct {
	key string
}

func (input headerInput) read(request *lazyRequest) reflect.Value {
	return reflect.ValueOf(request.httpRequest.Header.Get(input.key))
}

type pathInput struct{}

func (input pathInput) read(request *lazyRequest) reflect.Value {
	return reflect.ValueOf(request.httpRequest.URL.Path)
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
