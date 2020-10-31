package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8041", "http service address")

func main() {
	flag.Parse()
	initBackends()
	initServer()
}

func initServer() {
	http.Handle("/", http.HandlerFunc(cGPHandleRequest))
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
