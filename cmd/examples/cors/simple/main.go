package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	// Parse the server address.
	addr := flag.String("addr", ":9090", "Server address")
	flag.Parse()

	// Log the starting server message.
	log.Printf("starting server on %s", *addr)

	// Start a static file server.
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
