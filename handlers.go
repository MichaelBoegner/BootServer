package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	IsChirpyRed  bool   `json:"is_chirpy_red,omitemtpy"`
}

type acceptedVals struct {
	Body             string `json:"body"`
	Password         string `json:"password"`
	Email            string `json:"email"`
	ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	Event            string `json:"event,omitempty"`
	Data             data   `json:"data,omitempty"`
}

type data struct {
	UserID int `json:"user_id,omitempty"`
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
		params, err := getParams(r, w)
		if err != nil {
			log.Printf("\nError: %v", err)
		}

		tokenString, err := getHeaderToken(r)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}

		authorID, token := verifyToken(tokenString, w)
		if !token {

			respondWithError(w, 401, "Unauthorized")
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
				Body:     joinedBody,
				AuthorID: authorID,
			}

			_, err := cfg.db.CreateChirp(payload.Body, authorID)
			if err != nil {
				log.Printf("Chirp not created by CreateChirp(): %v", err)
			}

			payload.Id = len(cfg.db.DatabaseStructure.Chirps)

			respondWithJSON(w, 201, payload)
			return
		} else {
			respondWithError(w, 400, "Body must be 140 characters or less")
			return
		}
	case http.MethodGet:
		fmt.Printf("\nMethodGet Section Firing\n")
		path := r.URL.Path
		split_path := strings.Split(path, "/")
		author_id := r.URL.Query().Get("author_id")

		if author_id != "" {
			authorID, err := strconv.Atoi(author_id)
			fmt.Printf("\nauthorID converted to INT: %v\n", authorID)
			if err != nil {
				respondWithError(w, 500, "author_id did not convert to int")
				return
			}
			chirps, err := cfg.db.GetChirpsByAuthor(authorID)
			fmt.Printf("\nChirps returned by author id: %v\n", chirps)
			if err != nil {
				log.Printf("\nError: %v", err)
				respondWithError(w, 401, "No chirps by that author")
				return
			}
			var payload []returnVals
			for k, v := range chirps {
				payload = append(payload, returnVals{Id: k, Body: v.Body, AuthorID: v.AuthorID})
			}
			fmt.Printf("\npayload built: %v\n", payload)
			if r.URL.Query().Get("sort") == "desc" {
				sort.SliceStable(payload, func(i, j int) bool { return payload[i].Id > payload[j].Id })
			} else {
				sort.SliceStable(payload, func(i, j int) bool { return payload[i].Id < payload[j].Id })
			}

			respondWithJSON(w, 200, payload)
			return

		} else if len(split_path) == 4 && split_path[3] != "" {
			dynamic_id, _ := strconv.Atoi(split_path[3])
			chirp, err := cfg.db.GetChirp(dynamic_id)
			if err != nil {
				respondWithError(w, 404, "Invalid ID")
				return
			}
			payload := returnVals{Id: dynamic_id, Body: chirp.Body}
			respondWithJSON(w, 200, payload)
			return
		} else if len(split_path) == 4 && split_path[3] == "" {
			fmt.Printf("\nelse if firing: %v\n", split_path)
			var payload []returnVals
			for k, v := range cfg.db.DatabaseStructure.Chirps {
				payload = append(payload, returnVals{Id: k, Body: v.Body, AuthorID: v.AuthorID})
			}

			if r.URL.Query().Get("sort") == "desc" {
				fmt.Printf("\nSORT Desc firing")
				sort.SliceStable(payload, func(i, j int) bool { return payload[i].Id > payload[j].Id })
			} else {
				sort.SliceStable(payload, func(i, j int) bool { return payload[i].Id < payload[j].Id })
			}
			respondWithJSON(w, 200, payload)
			return
		} else {
			fmt.Printf("\nElse firing for some reason\n")
			respondWithError(w, 400, "Invalid path")
			return
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

			authorID, token := verifyToken(tokenString, w)
			if !token {
				respondWithError(w, http.StatusForbidden, "Unauthorized")
				return
			}

			deleted := cfg.db.DeleteChirp(dynamic_id, authorID)
			if !deleted {

				respondWithError(w, http.StatusForbidden, "Unauthorized")
				return
			}

			payload := &returnVals{}
			respondWithJSON(w, 204, payload)
			return
		}

		respondWithError(w, 403, "Unauthorized")
	}
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	params, err := getParams(r, w)
	if err != nil {
		log.Printf("\nError: %v", err)
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
		payload.IsChirpyRed = cfg.db.DatabaseStructure.Users[payload.Id].IsChirpyRed
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
			Id:          id,
			Email:       user.Email,
			IsChirpyRed: cfg.db.DatabaseStructure.Users[id].IsChirpyRed,
		}
		respondWithJSON(w, 200, payload)
	}
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	jwtSecret := cfg.jwt

	params, err := getParams(r, w)
	if err != nil {
		log.Printf("\nError: %v", err)
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
		IsChirpyRed:  cfg.db.DatabaseStructure.Users[id].IsChirpyRed,
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

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	token, err := getHeaderToken(r)
	if err != nil {
		respondWithError(w, 401, "")
		return
	}
	api_key := os.Getenv("API_KEY")
	if token != api_key {
		respondWithError(w, 401, "")
		return
	}

	params, err := getParams(r, w)
	if err != nil {
		log.Printf("\nError: %v", err)
	}
	fmt.Printf("\nParams: %v", params)
	if params.Event != "user.upgraded" {
		respondWithJSON(w, 204, "")
		return
	}

	_, err = cfg.db.AddUpgradeBadge(params.Data.UserID)
	if err != nil {
		respondWithError(w, 404, "")
		return
	}
	payload := &returnVals{}

	respondWithJSON(w, 204, payload)
	return

}

func getHeaderToken(r *http.Request) (string, error) {
	tokenParts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(tokenParts) < 2 {
		err := errors.New("Authoization header is malformed")
		return "", err
	}
	return tokenParts[1], nil
}

func getParams(r *http.Request, w http.ResponseWriter) (acceptedVals, error) {
	decoder := json.NewDecoder(r.Body)
	params := acceptedVals{}
	err := decoder.Decode(&params)
	if err != nil {

		respondWithError(w, 500, "Error decoding params")
		return params, err
	}

	return params, nil
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

		respondWithError(w, 403, "Unauthorized")
		return 0, false
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		log.Fatalf("ID not converted from string to int: %v", err)
	}
	return id, true
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Add("Content-Type", "application/json")
	respBody := returnVals{
		Error: msg,
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}
