package main

import (
	"encoding/json"
	"log"
	"net/http"
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
		Error string `json:"error"`
		Valid bool   `json:"valid"`
	}

	if len(params.Body) <= 140 {
		respBody := returnVals{
			Valid: true,
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
