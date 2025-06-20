package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Chirp string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if len(params.Chirp) > 141 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	bad_words := []string{"kerfuffle", "sharbert", "fornax"}
	cleaned_chirp := cleanChirp(bad_words, params.Chirp)

	type returnValue struct {
		CleanedBody string `json:"cleaned_body"`
	}

	respondWithJson(w, http.StatusOK, returnValue{
		CleanedBody: cleaned_chirp,
	})
}

func cleanChirp(bad_words []string, chirp string) string {
	tokens := strings.Split(chirp, " ")
	for i, token := range tokens {
		for _, bw := range bad_words {
			if strings.ToLower(token) == bw {
				tokens[i] = "****"
			}
		}
	}

	return strings.Join(tokens, " ")
}
