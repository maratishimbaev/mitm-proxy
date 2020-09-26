package models

type Request struct {
	Host  string
	Port  string
	IsSSL bool
	Head  string
	Body  string
}
