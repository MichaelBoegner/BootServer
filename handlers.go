package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type returnVals struct {
	Error string `json:"error,omitempty"`
	Id    int    `json:"id,omitempty"`
	Body  string `json:"body,omitempty"`
	Email string `json:"email,omitempty"`
	Token string `json:"token,omitempty"`
}

type acceptedVals struct {
	Body             string `json:"body"`
	Password         string `json:"password"`
	Email            string `json:"email"`
	ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var err error
	cfg.db.DatabaseStructure, err = cfg.db.LoadDB()
	if err != nil {
		log.Printf("Error loading database: %s", err)
	}

	switch r.Method {
	case http.MethodPost:
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

			payload := &returnVals{
				Body: joinedBody,
			}

			_, err := cfg.db.CreateChirp(payload.Body)
			if err != nil {
				log.Printf("Chirp not created by CreateChirp(): %v", err)
			}

			payload.Id = len(cfg.db.DatabaseStructure.Chirps)

			respondWithJSON(w, 201, payload)
		} else {
			respondWithError(w, 400, "Body must be 140 characters or less")
		}

	case http.MethodGet:
		path := r.URL.Path
		split_path := strings.Split(path, "/")

		if len(split_path) == 4 && split_path[3] != "" {
			dynamic_id, _ := strconv.Atoi(split_path[3])
			chirp, err := cfg.db.GetChirp(dynamic_id)
			if err != nil {
				respondWithError(w, 404, "Invalid ID")
				return
			}
			payload := returnVals{Id: dynamic_id, Body: chirp.Body}
			respondWithJSON(w, 200, payload)
		} else if len(split_path) == 4 && split_path[3] == "" {
			var payload []returnVals
			for k, v := range cfg.db.DatabaseStructure.Chirps {
				payload = append(payload, returnVals{Id: k, Body: v.Body})
			}
			sort.SliceStable(payload, func(i, j int) bool { return payload[i].Id < payload[j].Id })
			respondWithJSON(w, 200, payload)
		} else {
			respondWithError(w, 400, "Invalid path")
		}
	}
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := acceptedVals{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	switch r.Method {
	case http.MethodPost:
		user, err := cfg.db.CreateUser(params.Email, params.Password)
		if err != nil {
			log.Printf("Email parameter not valid: %s", err)
		}

		payload := &returnVals{
			Email: user.Email,
		}
		payload.Id = len(cfg.db.DatabaseStructure.Users)

		respondWithJSON(w, 201, payload)

	case http.MethodPut:
		headerAuth := r.Header.Get()
		token := ""
		user, err := cfg.db.UpdateUser(token)
		if err != nil {
			log.Printf("Email parameter not valid: %s", err)
		}

	}

}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	jwtSecret := cfg.jwt

	decoder := json.NewDecoder(r.Body)
	params := acceptedVals{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	user, id, token, err := cfg.db.GetUser(params.Email, params.Password, jwtSecret, params.ExpiresInSeconds)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
	}

	payload := &returnVals{
		Id:    id,
		Email: user.Email,
		Token: token,
	}
	fmt.Printf("PAYLOAD == %v", payload)
	respondWithJSON(w, 200, payload)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)

	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	log.Printf("Responding with status: %d", code)
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
