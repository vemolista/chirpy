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
	"github.com/vemolista/chirpy/v2/internal/auth"
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
		Body string `json:"body"`
	}

	type response struct {
		Chirp
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting Bearer token", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating JWT", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding JSON", err)
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
		UserID: userId,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
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
	s := r.URL.Query().Get("author_id")

	// very undry ¯\_(ツ)_/¯
	if s != "" {
		userId, err := uuid.Parse(s)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author uuid", err)
			return
		}

		data, err := cfg.db.ListChirpsForAuthor(r.Context(), userId)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error getting chirps for author", err)
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
	} else {
		data, err := cfg.db.ListChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error getting chirps", err)
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

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting token from header", err)
		return
	}

	userId, err := auth.ValidateJWT(accessToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating JWT", err)
		return
	}

	chirpId := r.PathValue("chirpId")
	parsedChirpId, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing ID of chirp", err)
		return
	}

	chirpData, err := cfg.db.GetChirp(r.Context(), parsedChirpId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp does not exist", err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Error getting chirp from db", err)
		return
	}

	if chirpData.UserID != userId {
		respondWithError(w, http.StatusForbidden, "Cannot delete chirps of other users", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpData.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting chirp from db", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
