package models

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

type Request struct {
	Id uint64
	Method string
	Path string
	Protocol string
	Headers string
	Body string
}

func FromHttpRequest(r *http.Request) *Request {
	headersJson, _ := json.Marshal(r.Header)

	body := new(bytes.Buffer)
	_, _ = body.ReadFrom(r.Body)

	return &Request{
		Method:   r.Method,
		Path:     r.RequestURI,
		Protocol: r.Proto,
		Headers:  string(headersJson),
		Body:     body.String(),
	}
}

func ToHttpRequest(r *Request) *http.Request {
	request, err := http.NewRequest(r.Method, r.Path, strings.NewReader(r.Body))
	if err != nil {
		return nil
	}

	request.Proto = r.Protocol

	var headers http.Header
	_ = json.Unmarshal([]byte(r.Headers), &headers)

	return request
}
