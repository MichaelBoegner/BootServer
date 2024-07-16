package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

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

	type returnVals struct {
		Error        string `json:"error"`
		Cleaned_body string `json:"cleaned_body"`
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

		respBody := returnVals{
			Cleaned_body: joinedBody,
		}

		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	} else {
		respBody := returnVals{
			Error: "Body is too long",
		}

		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(400)
		w.Write(data)
	}
}
