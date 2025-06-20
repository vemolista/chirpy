package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const PORT = ":8080"

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	cfg := apiConfig{}

	serveMux := http.NewServeMux()

	appHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", cfg.middlewareMetricsInc(appHandler))
	serveMux.HandleFunc("GET /api/healthz", healthHandler)
	serveMux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	serveMux.HandleFunc("POST /admin/reset", cfg.resetMetricsHandler)
	serveMux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	httpServer := http.Server{
		Handler: serveMux,
		Addr:    PORT,
	}

	fmt.Printf("Listening on port %v\n", PORT)
	httpServer.ListenAndServe()
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Chirp string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"error\":\"Something went wrong.\"}"))
		return
	}

	if len(params.Chirp) > 141 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"error\":\"Chirp is too long.\"}"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"valid\": true}"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Printf("error writing request body: %v\n", err)
	}
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	html := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, cfg.fileserverHits.Load())

	_, err := w.Write([]byte(html))
	if err != nil {
		fmt.Printf("error writing request body: %v\n", err)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

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
