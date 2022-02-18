package main

import (
	"fmt"
	"net/http"
	"os"
)

func response(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello from blue! I was called with URL %s\n", req.URL.Path)
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		fmt.Println("Could not detect PORT from environment variables")
		return
	}

	fmt.Printf("Starting blue on port %s\n", port)

	http.HandleFunc("/", response)

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
