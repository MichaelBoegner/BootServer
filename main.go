package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/michaelboegner/bootserver/internal/database"
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
	fmt.Printf("\n\nDATABASE FUNCTION NEWDB ==  %v, AND THIS IS ERROR == %v", db, err)

	databaseStructure, err := db.LoadDB()
	fmt.Printf("\n\nDATABASE STRUCTURE ==  %v, AND THIS IS ERROR == %v", databaseStructure, err)

	chirp, err := db.CreateChirp("This is a chirp")
	fmt.Printf("\n\nCHIRP ==  %v, AND THIS IS ERROR == %v", chirp, err)

	databaseStructure, err = db.LoadDB()
	fmt.Printf("\n\nDATABASE STRUCTURE.CHIRPS AFTER WRITING CHIRP ==  %v, AND THIS IS ERROR == %v", databaseStructure.Chirps[0], err)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
