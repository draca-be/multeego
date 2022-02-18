package main

import (
	"flag"
	"fmt"
	"net/http"
)

func response(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello from red! I was called with URL %s\n", req.URL.Path)
}

func main() {
	portPtr := flag.Int("port", 8080,  "port number")

	flag.Parse()

	fmt.Printf("Starting red on port %d\n", *portPtr)

	http.HandleFunc("/", response)

	http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), nil)
}
