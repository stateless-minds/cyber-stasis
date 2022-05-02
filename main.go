package main

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// The main function is the entry point where the app is configured and started.
// It is executed in 2 different environments: A client (the web browser) and a
// server.
func main() {
	app.Route("/", &pubsub{})
	app.RunWhenOnBrowser()
	initServer()
}
