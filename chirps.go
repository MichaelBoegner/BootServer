package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/michaelboegner/bootserver/internal/database"
	"k8s.io/kube-openapi/pkg/validation/errors"
)

type returnVals struct {
	Error string `json:"error,omitempty"`
	Id    int    `json:"id,omitempty"`
	Body  string `json:"body,omitempty"`
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request, db *database.DB) {
	database, err := db.LoadDB()
	if err != nil {
		errors.Error("Database not loading in handlerChirps")
	}

	fmt.Printf("\n\nTHIS IS DATABASE == %v", database)
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
