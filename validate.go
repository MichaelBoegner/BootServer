package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type returnVals struct {
	Error        string `json:"error,omitempty"`
	Cleaned_body string `json:"cleaned_body,omitempty"`
}

func (cfg *apiConfig) handlerJSON(w http.ResponseWriter, r *http.Request) {
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
			Cleaned_body: joinedBody,
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
