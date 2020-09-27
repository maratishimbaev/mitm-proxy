package proxyHttp

import (
	"github.com/kataras/golog"
	"io"
	"mitm-proxy/app/models"
	"mitm-proxy/app/proxy/interfaces"
	"net"
	"net/http"
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
		h.handleHttps(w, r)
		return
	}
	h.handleHttp(w, r)
}

func (h *handler) handleHttp(w http.ResponseWriter, r *http.Request) {
	request := models.FromHttpRequest(r)

	err := h.useCase.CreateRequest(request)
	if err != nil {
		golog.Error(err.Error())
	}

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

func (h *handler) handleHttps(w http.ResponseWriter, r *http.Request) {
	request := models.FromHttpRequest(r)

	err := h.useCase.CreateRequest(request)
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

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()

	_, _ = io.Copy(destination, source)
}
