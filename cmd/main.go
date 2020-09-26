package main

import "mitm-proxy/app/server"

func main() {
	app := server.NewApp()
	app.Start()
}
