package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type returnVals struct {
	Error string `json:"error,omitempty"`
	Id    int    `json:"id,omitempty"`
	Body  string `json:"body,omitempty"`
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	type acceptedVals struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := acceptedVals{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(params.Body) <= 140 {
		splitBody := strings.Split(params.Body, " ")

		for i, word := range splitBody {
			word = strings.ToLower(word)
			if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
				splitBody[i] = "****"
			}
		}
		joinedBody := strings.Join(splitBody, " ")

		payload := returnVals{
			Body: joinedBody,
		}

		cfg.db.DatabaseStructure, err = cfg.db.LoadDB()

		if err != nil {
			log.Printf("Error loading database: %s", err)
		}
		fmt.Printf("\n\nDATABASE LOADED == %v", cfg.db.DatabaseStructure)

		_, err := cfg.db.CreateChirp(payload.Body)

		payload.Id = len(cfg.db.DatabaseStructure.Chirps) - 1

		fmt.Printf("\n\npayload== %v", payload)

		if err != nil {
			log.Printf("Chirp not created by CreateChirp(): %v", err)
		}

		respondWithJSON(w, 200, payload)

	} else {
		respondWithError(w, 400, "Body must be 140 characters or less")
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)

	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respBody := returnVals{
		Error: msg,
	}

	data, err := json.Marshal(respBody)

	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}
