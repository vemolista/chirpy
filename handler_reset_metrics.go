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
}
