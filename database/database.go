package database

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

type User struct {
	Password         []byte    `json:"password"`
	Email            string    `json:"email"`
	ExpiresInSeconds int       `json:"expires_in_seconds"`
	RefreshToken     string    `json:"refresh_token"`
	TokenExpiry      time.Time `json:"token_expiry"`
	IsChirpyRed      bool      `json:"is_chirpy_red"`
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

func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	chirp := Chirp{
		Body:     body,
		AuthorID: authorID,
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

func (db *DB) GetChirpsByAuthor(authorId int) (map[int]Chirp, error) {
	fmt.Printf("\nGetGhirpsByAuthor firing\n")
	db.mux.RLock()
	defer db.mux.RUnlock()

	var blankChirps map[int]Chirp
	chirps := make(map[int]Chirp)

	for i, chirp := range db.DatabaseStructure.Chirps {
		if chirp.AuthorID == authorId {
			fmt.Printf("\nIf firing: %v, %v\n", chirp, i)
			chirps[i] = chirp
			fmt.Printf("\nThis shouldn't be firing: %v", chirps, i)
		}
	}
	fmt.Printf("\nchirps appended found by author_id: %v\n", chirps)
	if chirps == nil {
		err := errors.New("No chirps returned by author_id")
		return blankChirps, err
	}

	return chirps, nil
}

func (db *DB) DeleteChirp(id, authorID int) bool {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if db.DatabaseStructure.Chirps[id].AuthorID == authorID {
		delete(db.DatabaseStructure.Chirps, id)
	} else {
		return false
	}
	err := writeFile(db)
	if err != nil {
		log.Printf("\nError: %v", err)
	}

	var blankChirp Chirp
	if db.DatabaseStructure.Chirps[id] != blankChirp {
		return false
	}
	return true
}

func (db *DB) CreateUser(email, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	user := User{
		Email:       email,
		IsChirpyRed: false,
	}

	var err error
	user.Password, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return user, err
	}

	nextID := len(db.DatabaseStructure.Users) + 1
	db.DatabaseStructure.Users[nextID] = user
	err = writeFile(db)
	if err != nil {
		log.Fatalf("Not writing to database: %v", err)
	}

	return user, nil
}

func (db *DB) LoginUser(email, password, jwtSecret string, expires int) (User, int, string, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	var (
		User User
		id   int
	)

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
	if id == 0 {
		return User, id, "", errors.New("Email not found")
	}

	token, err := createJWTToken(id, expires)
	refreshToken, tokenExpiry, err := createRefreshToken()

	User.RefreshToken = refreshToken
	User.TokenExpiry = tokenExpiry

	db.DatabaseStructure.Users[id] = User
	err = writeFile(db)
	if err != nil {
		log.Fatalf("Not writing to database: %v", err)
	}

	return User, id, token, nil
}

func (db *DB) GetUserbyRefreshToken(refreshToken string) (User, string, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	var (
		User User
		id   int
	)

	found := false

	for i, user := range db.DatabaseStructure.Users {
		if user.RefreshToken == refreshToken {
			User = user
			id = i
			found = true
		}
	}
	if !found {
		err := errors.New("refresh token not found")
		return User, "", err
	}

	if !User.TokenExpiry.After(time.Now()) {
		err := errors.New("refresh token not found")
		return User, "", err
	}

	token, err := createJWTToken(id, 0)
	if err != nil {
		log.Printf("JWT token not created: %v", err)
	}

	return User, token, nil
}

func (db *DB) UpdateUser(password, email string, id int) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	User := db.DatabaseStructure.Users[id]

	fmt.Printf("\nUser before assignments: %v\n", User)
	fmt.Printf("\nDB user by id: %v\n", db.DatabaseStructure.Users[id])
	var err error
	User.Email = email
	User.Password, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		log.Printf("Unable to create password: %v", err)
		return User, err
	}

	db.DatabaseStructure.Users[id] = User
	err = writeFile(db)
	if err != nil {
		log.Fatalf("Not writing to database: %v", err)
	}

	return User, nil
}

func (db *DB) AddUpgradeBadge(id int) (User, error) {
	var (
		User    User
		foundID int
		found   bool
	)

	for i, user := range db.DatabaseStructure.Users {
		if id == i {
			User = user
			foundID = i
			found = true
		}
	}
	if !found {
		err := errors.New("User with user_id does not exist.")
		return User, err
	}

	User.IsChirpyRed = true
	db.DatabaseStructure.Users[foundID] = User

	err := writeFile(db)
	if err != nil {
		log.Printf("\nDatabase not updated with upgraded badge")
	}

	return User, nil
}

func (db *DB) RevokeRefreshToken(user User) error {
	var (
		id       int
		copyUser User
	)

	for i, User := range db.DatabaseStructure.Users {
		if User.RefreshToken == user.RefreshToken {
			id = i
			copyUser = User
		}
	}

	copyUser.RefreshToken = ""

	db.DatabaseStructure.Users[id] = copyUser

	err := writeFile(db)
	if err != nil {
		log.Printf("Error writing to database: %v", err)
		return err
	}

	return nil
}

func createJWTToken(id, expires int) (string, error) {
	var (
		key []byte
		t   *jwt.Token
	)

	jwtSecret := os.Getenv("JWT_SECRET")
	now := time.Now()
	if expires == 0 {
		expires = 3600
	}
	expiresAt := time.Now().Add(time.Duration(expires) * time.Second)
	key = []byte(jwtSecret)
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   strconv.Itoa(id),
	}
	t = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(key)
	if err != nil {
		log.Fatalf("Bad SignedString: %s", err)
		return "", err
	}

	return s, nil
}

func createRefreshToken() (string, time.Time, error) {
	refreshLength := 32
	refreshBytes := make([]byte, refreshLength)
	_, err := rand.Read([]byte(refreshBytes))
	if err != nil {
		return "", time.Time{}, err
	}
	refreshToken := hex.EncodeToString(refreshBytes)
	expiry := time.Now().Add(time.Duration(24*60) * time.Hour)

	return refreshToken, expiry, nil
}

func writeFile(db *DB) error {
	marshaledData, err := json.Marshal(db.DatabaseStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, marshaledData, 0666)
	if err != nil {
		return err
	}
	return nil
}
