package main

import (
	"github.com/labstack/gommon/log"
	"net/http"
)

func main() {
	p := NewProxy()

	log.Info("Server started\n")
	log.Fatal(http.ListenAndServe(":8000", p))
}
