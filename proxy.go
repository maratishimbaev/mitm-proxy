package main

import (
	"context"
	"fmt"
	"github.com/kataras/golog"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type Proxy struct {
	http.Handler
	col *mongo.Collection
}

func NewProxy(col *mongo.Collection) *Proxy {
	return &Proxy{
		col: col,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleHTTPS(w, r)
		return
	}
	p.handleHTTP(w, r)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()

	_, _ = io.Copy(destination, source)
}

func (p *Proxy) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ":")
	head := fmt.Sprintf("%s", r.Method)
	body, _ := ioutil.ReadAll(r.Body)

	request := NewRequest(
		host[0],
		host[1],
		true,
		head,
		string(body),
	)
	golog.Infof("request: %s", request)

	_, err := p.col.InsertOne(context.TODO(), request)
	if err != nil {
		golog.Error(err.Error())
	}

	destinationConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking error", http.StatusInternalServerError)
		return
	}
	sourceConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go transfer(destinationConn, sourceConn)
	go transfer(sourceConn, destinationConn)
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
