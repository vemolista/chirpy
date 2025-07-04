package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/vemolista/chirpy/v2/internal/database"
)

const PORT = ":8080"

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
	polkaKey       string
}

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	dbConnection, err := sql.Open("postgres", dbUrl)

	if err != nil {
		panic("Error opening a database connection")
	}

	dbQueries := database.New(dbConnection)

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
		polkaKey:       polkaKey,
	}

	serveMux := http.NewServeMux()

	appHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", cfg.middlewareMetricsInc(appHandler))

	serveMux.HandleFunc("GET /api/healthz", healthHandler)
	serveMux.HandleFunc("POST /api/chirps", cfg.createChirpHandler)
	serveMux.HandleFunc("GET /api/chirps", cfg.listChirpsHandler)
	serveMux.HandleFunc("GET /api/chirps/{chirpId}", cfg.getChirpHandler)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpId}", cfg.deleteChirpHandler)
	serveMux.HandleFunc("POST /api/users", cfg.createUserHandler)
	serveMux.HandleFunc("PUT /api/users", cfg.updateUserHandler)
	serveMux.HandleFunc("POST /api/login", cfg.loginHandler)
	serveMux.HandleFunc("POST /api/refresh", cfg.refreshHandler)
	serveMux.HandleFunc("POST /api/revoke", cfg.revokeHandler)

	serveMux.HandleFunc("POST /api/polka/webhooks", cfg.polkaWebhookHandler)

	serveMux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	serveMux.HandleFunc("POST /admin/reset", cfg.resetMetricsHandler)

	httpServer := http.Server{
		Handler: serveMux,
		Addr:    PORT,
	}

	fmt.Printf("Listening on port %v\n", PORT)
	httpServer.ListenAndServe()
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}
