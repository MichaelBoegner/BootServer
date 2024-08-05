package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/michaelboegner/bootserver/database"
)

type apiConfig struct {
	fileserverHits int
	db             *database.DB
	jwt            string
}

func main() {
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")

	db, err := database.NewDB("database/database.json")
	if err != nil {
		log.Fatalf("Failed to initialize database due to following errror: %s", err)
	}

	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
		db:             db,
		jwt:            jwtSecret,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/api/healthz", handlerReadiness)
	mux.HandleFunc("/api/reset", apiCfg.handlerReset)

	mux.HandleFunc("/api/chirps", apiCfg.handlerChirps)
	mux.HandleFunc("/api/users", apiCfg.handlerUsers)
	mux.HandleFunc("/api/login", apiCfg.handlerLogin)
	mux.HandleFunc("/api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("/api/revoke", apiCfg.handlerRevoke)
	mux.HandleFunc("/api/polka/webhooks", apiCfg.handlerWebhooks)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
