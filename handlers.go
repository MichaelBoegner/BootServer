package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type returnVals struct {
	Error        string `json:"error,omitempty"`
	Id           int    `json:"id,omitempty"`
	Body         string `json:"body,omitempty"`
	Email        string `json:"email,omitempty"`
	Token        string `json:"token,omitempty"`
	Password     []byte `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	AuthorID     int    `json:"author_id,omitempty"`
}

type acceptedVals struct {
	Body             string `json:"body"`
	Password         string `json:"password"`
	Email            string `json:"email"`
	ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
}

type MyCustomClaims struct {
	jwt.RegisteredClaims
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
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

		tokenString, err := getHeaderToken(r)
		if err != nil {
			log.Printf("Error: %v", err)
		}

		authorID, token := verifyToken(tokenString, w)
		if !token {
			respondWithError(w, 401, "Unauthorized")
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
				Body:     joinedBody,
				AuthorID: authorID,
			}

			_, err := cfg.db.CreateChirp(payload.Body, authorID)
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
				payload = append(payload, returnVals{Id: k, Body: v.Body, AuthorID: v.AuthorID})
			}
			sort.SliceStable(payload, func(i, j int) bool { return payload[i].Id < payload[j].Id })
			respondWithJSON(w, 200, payload)
		} else {
			respondWithError(w, 400, "Invalid path")
		}
	case http.MethodDelete:
		path := r.URL.Path
		split_path := strings.Split(path, "/")

		if len(split_path) == 4 && split_path[3] != "" {
			dynamic_id, err := strconv.Atoi(split_path[3])
			if err != nil {
				respondWithError(w, 404, "Invalid ID")
				return
			}

			tokenString, err := getHeaderToken(r)
			if err != nil {
				log.Printf("Error: %v", err)
			}

			_, token := verifyToken(tokenString, w)

			if !token {
				respondWithError(w, 403, "Unauthorized")
				return
			}

			deleted := cfg.db.DeleteChirp(dynamic_id)
			if !deleted {
				respondWithError(w, 401, "Chirp not deleted")
			}

			payload := &returnVals{}
			respondWithJSON(w, 204, payload)

		}
	}
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
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
			respondWithError(w, 400, "Bad Request")
			return
		}

		payload := &returnVals{
			Email: user.Email,
		}
		payload.Id = len(cfg.db.DatabaseStructure.Users)

		respondWithJSON(w, 201, payload)

	case http.MethodPut:
		tokenString, err := getHeaderToken(r)
		if err != nil {
			log.Printf("Error: %v", err)
		}

		id, token := verifyToken(tokenString, w)
		if !token {
			respondWithError(w, 401, "Unauthorized")
		}

		user, err := cfg.db.UpdateUser(params.Password, params.Email, id)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			respondWithError(w, 500, "Internal Server Error")
			return
		}

		payload := &returnVals{
			Id:    id,
			Email: user.Email,
		}

		respondWithJSON(w, 200, payload)

	}

}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	jwtSecret := cfg.jwt

	decoder := json.NewDecoder(r.Body)
	params := acceptedVals{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	user, id, token, err := cfg.db.LoginUser(params.Email, params.Password, jwtSecret, params.ExpiresInSeconds)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
	}

	payload := &returnVals{
		Id:           id,
		Email:        user.Email,
		Token:        token,
		RefreshToken: user.RefreshToken,
	}

	respondWithJSON(w, 200, payload)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := getHeaderToken(r)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	_, token, err := cfg.db.GetUserbyRefreshToken(refreshTokenString)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
	}

	payload := &returnVals{
		Token: token,
	}
	respondWithJSON(w, 200, payload)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := getHeaderToken(r)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	user, _, err := cfg.db.GetUserbyRefreshToken(refreshTokenString)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
	}

	err = cfg.db.RevokeRefreshToken(user)
	if err != nil {
		log.Printf("Token not revoked: %v", err)
	}

	payload := &returnVals{}
	respondWithJSON(w, 204, payload)
}

func getHeaderToken(r *http.Request) (string, error) {
	tokenParts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(tokenParts) < 2 {
		err := errors.New("Authoization header is malformed")
		log.Fatal("\nError: %v", err)
		return "", err
	}
	return tokenParts[1], nil
}

func verifyToken(tokenString string, w http.ResponseWriter) (int, bool) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT secret is not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return 0, true
	}
	if token == nil {
		log.Fatal("Token parsing resulted in nil token")
	}

	idString, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, 500, "Internal Server Error")
		return 0, false
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		log.Fatalf("ID not converted from string to int: %v", err)
	}
	return id, true
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Add("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		respondWithError(w, 500, "Internal Server Error")
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
