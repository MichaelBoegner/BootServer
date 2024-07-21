package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/michaelboegner/internal/database"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)

	db, err := database.NewDB("internal/database/database.json")
	fmt.Printf("\n\n1. DATABASE FUNCTION NEWDB ==  %v, AND THIS IS ERROR == %v", db, err)

	// databaseStructure, err := db.LoadDB()
	// fmt.Printf("\n\n2. DATABASE STRUCTURE ==  %v, AND THIS IS ERROR == %v", databaseStructure, err)

	// chirp, err := db.CreateChirp("This is a chirp")
	// fmt.Printf("\n\n3. CHIRP ==  %v, AND THIS IS ERROR == %v", chirp, err)

	// databaseStructure, err = db.LoadDB()
	// fmt.Printf("\n\n4. DATABASE STRUCTURE.CHIRPS AFTER WRITING CHIRP ==  %v, AND THIS IS ERROR == %v", databaseStructure, err)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
