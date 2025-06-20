package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	cfg.fileserverHits.Store(0)
	msg := "Hits reset to 0"

	_, err := w.Write([]byte(msg))
	if err != nil {
		fmt.Printf("error writing request body: %v\n", err)
	}
}
