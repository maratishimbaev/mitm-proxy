package main

type Request struct {
	Host string
	Port string
	IsSSL bool
	Head string
	Body string
}

func NewRequest(host string, port string, isSSL bool, head string, body string) *Request {
	return &Request{
		Host: host,
		Port: port,
		IsSSL: isSSL,
		Head: head,
		Body: body,
	}
}
