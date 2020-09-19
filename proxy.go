package main

import (
	"io"
	"net/http"
)

type Proxy struct {
	http.Handler
}

func NewProxy() *Proxy {
	return &Proxy{}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleHTTPS(w, r)
		return
	}
	p.handleHTTP(w, r)
}

func (p *Proxy) handleHTTPS(w http.ResponseWriter, r *http.Request) {

}

func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer res.Body.Close()

	for header, value := range res.Header {
		w.Header()[header] = value
	}
	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
}
