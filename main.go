package main

import (
	"log"
	"net/http"

	"github.com/michaelboegner/bootserver/database"
)

type apiConfig struct {
	fileserverHits int
	db             *database.DB
}

func main() {
	db, err := database.NewDB("database/database.json")
	if err != nil {
		log.Fatalf("Failed to initialize database due to following errror: %s", err)
	}

	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
		db:             db,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)
	mux.HandleFunc("/api/chirps/", apiCfg.handlerChirps)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
