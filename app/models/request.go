package models

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
)

var id uint64

type Request struct {
	Id       uint64 `json:"id"`
	Method   string `json:"method"`
	Host     string `json:"host"`
	Path     string `json:"path"`
	Protocol string `json:"protocol"`
	Headers  string `json:"headers"`
	Body     string `json:"body"`
}

func FromHttpRequest(r *http.Request) *Request {
	headersJson, _ := json.Marshal(r.Header)

	body := new(bytes.Buffer)
	_, _ = body.ReadFrom(r.Body)

	return &Request{
		Id:       atomic.AddUint64(&id, 1),
		Method:   r.Method,
		Host:     r.Host,
		Path:     r.URL.Path,
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
	request.URL.Path = r.Path

	request.URL.Host = r.Host
	request.Host = r.Host

	var headers http.Header
	_ = json.Unmarshal([]byte(r.Headers), &headers)
	for key, values := range headers {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	return request
}
