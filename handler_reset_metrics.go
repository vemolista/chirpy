package main

import (
	"net/http"
)

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	type returnValue struct {
		Body string `json:"body"`
	}

	respondWithJson(w, http.StatusOK, returnValue{
		Body: "Hits reset to 0",
	})

	// w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	// w.WriteHeader(http.StatusOK)

	// cfg.fileserverHits.Store(0)
	// msg := "Hits reset to 0"

	// _, err = w.Write([]byte(msg))
	// if err != nil {
	// 	fmt.Printf("error writing request body: %v\n", err)
	// }
}
