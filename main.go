package main

import (
	"fmt"
	"net/http"
)

const PORT = ":8080"

func main() {
	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir(".")))

	httpServer := http.Server{
		Handler: serveMux,
		Addr:    PORT,
	}

	fmt.Printf("Listening on port %v\n", PORT)
	httpServer.ListenAndServe()
}
