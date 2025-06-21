package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vemolista/chirpy/v2/internal/database"
)

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserId    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	type response struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if len(params.Body) > 141 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	bad_words := []string{"kerfuffle", "sharbert", "fornax"}
	cleaned_chirp := cleanChirp(bad_words, params.Body)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned_chirp,
		UserID: params.UserId,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	respondWithJson(w, http.StatusCreated, response{
		Chirp: Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserId:    chirp.UserID,
			Body:      chirp.Body,
		},
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

func (cfg *apiConfig) listChirpsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := cfg.db.ListChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	response := []Chirp{}
	for _, item := range data {
		chirp := Chirp{
			Id:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			UserId:    item.UserID,
			Body:      item.Body,
		}

		response = append(response, chirp)
	}

	respondWithJson(w, http.StatusOK, response)
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpId")

	parsedId, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing ID of chirp", err)
		return
	}

	data, err := cfg.db.GetChirp(r.Context(), parsedId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("No chirp with Id %s", id), err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	respondWithJson(w, http.StatusOK, Chirp{
		Id:        data.ID,
		UserId:    data.UserID,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
		Body:      data.Body,
	})
}
