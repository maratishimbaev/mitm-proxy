package proxyHttp

import (
	"crypto/tls"
	"errors"
	"github.com/kataras/golog"
	certificate "mitm-proxy/app/cert"
	"mitm-proxy/app/models"
	"mitm-proxy/app/proxy/interfaces"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

type handler struct {
	http.Handler
	useCase       proxyInterfaces.ProxyUseCase
	ServerConfig  *tls.Config
	ClientConfig  *tls.Config
	FlushInterval time.Duration
}

func NewHandler(useCase proxyInterfaces.ProxyUseCase) *handler {
	return &handler{useCase: useCase}
}

func (h *handler) Wrap(upstream http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "" {
			r.URL.Scheme = "https"
		}

		request := models.FromHttpRequest(r)

		golog.Infof("#%d %s %s%s", request.Id, request.Method, request.Host, request.Path)

		err := h.useCase.CreateRequest(request)
		if err != nil {
			golog.Error("can't create request")
		}

		upstream.ServeHTTP(w, r)
	})
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		h.handleHttps(w, r)
		return
	}

	rp := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Host = r.Host
			r.URL.Scheme = "http"
		},
		FlushInterval: h.FlushInterval,
	}
	r.URL.Host = "http"
	h.Wrap(rp).ServeHTTP(w, r)
}

func (h *handler) handleHttps(w http.ResponseWriter, r *http.Request) {
	hostName, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		golog.Errorf("no host: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	cert, err := certificate.GetCert(hostName)
	if err != nil {
		golog.Errorf("get cert err: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	var serverConn *tls.Conn

	serverConfig := new(tls.Config)
	if h.ServerConfig != nil {
		*serverConfig = *h.ServerConfig
	}
	serverConfig.Certificates = []tls.Certificate{*cert}
	serverConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		clientConfig := new(tls.Config)
		if h.ClientConfig != nil {
			*clientConfig = *h.ClientConfig
		}
		clientConfig.ServerName = hello.ServerName
		serverConn, err = tls.Dial("tcp", r.Host, clientConfig)
		if err != nil {
			golog.Errorf("dial tcp err: %s", err.Error())
			return nil, err
		}
		return certificate.GetCert(hostName)
	}

	clientConn, err := handshake(w, serverConfig)
	if err != nil {
		golog.Errorf("handshake err: %s", err.Error())
		return
	}
	defer clientConn.Close()
	if serverConn == nil {
		golog.Error("client conn is nil")
		return
	}
	defer serverConn.Close()

	dialer := &oneShotDialer{conn: serverConn}
	rp := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Host = r.Host
			r.URL.Scheme = "https"
		},
		Transport:     &http.Transport{DialTLS: dialer.Dial},
		FlushInterval: h.FlushInterval,
	}

	ch := make(chan int)
	wc := &onCloseConn{clientConn, func() { ch <- 0 }}
	http.Serve(&oneShotListener{wc}, h.Wrap(rp))
	<-ch
}

func handshake(w http.ResponseWriter, config *tls.Config) (net.Conn, error) {
	raw, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return nil, err
	}

	okHeader := []byte("HTTP/1.1 200 OK\r\n\r\n")
	if _, err = raw.Write(okHeader); err != nil {
		raw.Close()
		return nil, err
	}

	conn := tls.Server(raw, config)
	err = conn.Handshake()
	if err != nil {
		conn.Close()
		raw.Close()
		return nil, err
	}
	return conn, nil
}

type oneShotDialer struct {
	conn net.Conn
	mu   sync.Mutex
}

func (d *oneShotDialer) Dial(network, address string) (net.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.conn == nil {
		return nil, errors.New("conn is nil")
	}
	conn := d.conn
	d.conn = nil
	return conn, nil
}

type onCloseConn struct {
	net.Conn
	f func()
}

func (c *onCloseConn) Close() error {
	if c.f != nil {
		c.f()
		c.f = nil
	}
	return c.Conn.Close()
}

type oneShotListener struct {
	conn net.Conn
}

func (l *oneShotListener) Accept() (net.Conn, error) {
	if l.conn == nil {
		return nil, errors.New("conn is nil")
	}
	conn := l.conn
	l.conn = nil
	return conn, nil
}

func (l *oneShotListener) Close() error {
	return nil
}

func (l *oneShotListener) Addr() net.Addr {
	return l.conn.LocalAddr()
}
