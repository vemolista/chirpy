package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/vemolista/chirpy/v2/internal/auth"
)

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to get refresh token from Authorization header", err)
		return
	}

	tokenData, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Refresh token does not exist", err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Error getting refresh token from db", err)
		return
	}

	if tokenData.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is revoked", err)
		return
	}

	if tokenData.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is expired", err)
		return
	}

	newAccessToken, err := auth.MakeJWT(tokenData.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making making JWT", err)
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	respondWithJson(w, http.StatusOK, response{
		Token: newAccessToken,
	})
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to get bearer token from header", err)
		return
	}

	tokenData, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Refresh token does not exist", err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Error getting refresh token from db", err)
		return
	}

	if tokenData.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is expired", err)
		return
	}

	err = cfg.db.RevokeToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
