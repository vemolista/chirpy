package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vemolista/chirpy/v2/internal/auth"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresIn *int   `json:"expires_in_seconds"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding JSON", err)
		return
	}

	userData, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, userData.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	expires := time.Hour
	if params.ExpiresIn != nil {
		seconds := time.Duration(*params.ExpiresIn) * time.Second
		if seconds < expires {
			expires = seconds
		}
	}

	token, err := auth.MakeJWT(userData.ID, cfg.secret, expires)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating JWT", err)
	}

	type response struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	respondWithJson(w, http.StatusOK, response{
		Id:        userData.ID,
		CreatedAt: userData.CreatedAt,
		UpdatedAt: userData.UpdatedAt,
		Email:     userData.Email,
		Token:     token,
	})
}
