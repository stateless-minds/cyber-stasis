//go:build !wasm

// file: http_server.go

package main

import (
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func initServer() {
	withGz := gziphandler.GzipHandler(&app.Handler{
		Name:        "pubsub",
		Description: "A pubsub ipfs example",
		Styles: []string{
			"/web/app.css",
			"https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/css/bootstrap.min.css",
			"https://use.fontawesome.com/releases/v5.7.2/css/all.css",
		},
		Scripts: []string{
			"https://cdnjs.cloudflare.com/ajax/libs/jquery/3.2.1/jquery.min.js",
			"https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/js/bootstrap.bundle.min.js",
		},
	})
	http.Handle("/", withGz)

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
