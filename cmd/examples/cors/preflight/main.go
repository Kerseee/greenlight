package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":9090", "Server address")
	flag.Parse()

	log.Printf("Starting server on %s", *addr)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
