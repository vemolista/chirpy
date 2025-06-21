package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vemolista/chirpy/v2/internal/auth"
	"github.com/vemolista/chirpy/v2/internal/database"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	token, err := auth.MakeJWT(userData.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making refresh token", err)
		return
	}

	const sixtyDays = time.Hour * 24 * 60

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    userData.ID,
		ExpiresAt: time.Now().Add(sixtyDays),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating refresh token in db", err)
	}

	type response struct {
		Id           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	respondWithJson(w, http.StatusOK, response{
		Id:           userData.ID,
		CreatedAt:    userData.CreatedAt,
		UpdatedAt:    userData.UpdatedAt,
		Email:        userData.Email,
		Token:        token,
		RefreshToken: refreshToken,
	})
}
