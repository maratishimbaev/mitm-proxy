package proxyHttp

import (
	"fmt"
	"github.com/kataras/golog"
	"io"
	"io/ioutil"
	"mitm-proxy/app/models"
	"mitm-proxy/app/proxy/delivery/interfaces"
	"net"
	"net/http"
	"strings"
	"time"
)

type handler struct {
	http.Handler
	useCase proxyInterfaces.ProxyUseCase
}

func NewHandler(useCase proxyInterfaces.ProxyUseCase) *handler {
	return &handler{useCase: useCase}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		h.handleHTTPS(w, r)
		return
	}
	h.handleHTTP(w, r)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()

	_, _ = io.Copy(destination, source)
}

func (h *handler) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ":")
	head := fmt.Sprintf("%s", r.Method)
	body, _ := ioutil.ReadAll(r.Body)

	request := models.Request{
		Host:  host[0],
		Port:  host[1],
		IsSSL: true,
		Head:  head,
		Body:  string(body),
	}
	golog.Infof("request: %s", request)

	err := h.useCase.CreateRequest(&request)
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

func (h *handler) handleHTTP(w http.ResponseWriter, r *http.Request) {
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
