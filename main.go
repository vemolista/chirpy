package main

import (
	"fmt"
	"net/http"
)

const PORT = ":8080"

func main() {
	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	serveMux.HandleFunc("/healthz", healthHandler)

	httpServer := http.Server{
		Handler: serveMux,
		Addr:    PORT,
	}

	fmt.Printf("Listening on port %v\n", PORT)
	httpServer.ListenAndServe()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Printf("error writing request body: %v\n", err)
	}
}
