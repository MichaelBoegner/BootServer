package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path              string
	mux               *sync.RWMutex
	DatabaseStructure *DBStructure
}

type Chirp struct {
	Body string `json:"body"`
}

type User struct {
	Password         []byte `json:"password"`
	Email            string `json:"email"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

func NewDB(path string) (*DB, error) {
	chirpMap := make(map[int]Chirp)
	userMap := make(map[int]User)

	databaseStruct := &DBStructure{
		Chirps: chirpMap,
		Users:  userMap,
	}

	db := &DB{
		path:              path,
		mux:               &sync.RWMutex{},
		DatabaseStructure: databaseStruct,
	}

	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := os.ReadFile(db.path)
	if err != nil {
		db.DatabaseStructure = &DBStructure{
			Chirps: make(map[int]Chirp),
			Users:  make(map[int]User),
		}

		marshaledData, err := json.Marshal(db.DatabaseStructure)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(db.path, marshaledData, 0666)
		return db, err
	}

	err = json.Unmarshal(data, db.DatabaseStructure)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) LoadDB() (*DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	data, err := os.ReadFile(db.path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, db.DatabaseStructure)
	if err != nil {
		return nil, err
	}

	return db.DatabaseStructure, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	chirp := Chirp{
		Body: body,
	}

	nextID := len(db.DatabaseStructure.Chirps) + 1
	db.DatabaseStructure.Chirps[nextID] = chirp

	marshaledData, err := json.Marshal(db.DatabaseStructure)
	if err != nil {
		return chirp, err
	}

	err = os.WriteFile(db.path, marshaledData, 0666)
	if err != nil {
		return chirp, err
	}

	return chirp, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	chirp := db.DatabaseStructure.Chirps[id]
	if chirp.Body == "" {
		err := errors.New("Chirp not found")
		return chirp, err
	}
	return chirp, nil
}

func (db *DB) CreateUser(email, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	user := User{
		Email: email,
	}

	var err error
	user.Password, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return user, err
	}

	nextID := len(db.DatabaseStructure.Users) + 1
	db.DatabaseStructure.Users[nextID] = user

	marshaledData, err := json.Marshal(db.DatabaseStructure)
	if err != nil {
		return user, err
	}

	err = os.WriteFile(db.path, marshaledData, 0666)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (db *DB) GetUser(email, password, jwtSecret string, expires int) (User, int, string, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	var (
		key []byte
		t   *jwt.Token
	)
	now := time.Now()
	expiresAt := time.Now().Add(time.Duration(expires))
	key = []byte(jwtSecret)
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   strconv.Itoa(123),
	}
	t = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(key)
	if err != nil {
		log.Fatalf("Bad SignedString: %s", err)
	}

	var User User
	var id int
	for i, user := range db.DatabaseStructure.Users {
		if user.Email == email {
			err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
			if err != nil {
				return User, 0, "", err
			}
			User = user
			id = i
		}
	}

	return User, id, s, nil
}
