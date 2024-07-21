package database

import (
	"encoding/json"
	"os"
	"sync"
)

type DB struct {
	path              string
	mux               *sync.RWMutex
	databaseStructure *DBStructure
}

type Chirp struct {
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := os.ReadFile(db.path)
	if err != nil {
		db.databaseStructure = &DBStructure{Chirps: make(map[int]Chirp)}
		marshaledData, err := json.Marshal(db.databaseStructure)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(db.path, marshaledData, 0666)
		return db, err
	}

	db.databaseStructure = &DBStructure{}
	err = json.Unmarshal(data, db.databaseStructure)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	chirp := Chirp{Body: body}

	nextID := len(db.databaseStructure.Chirps) + 1
	db.databaseStructure.Chirps[nextID] = chirp

	marshaledChirp, err := json.Marshal(db.databaseStructure)
	if err != nil {
		return chirp, err
	}

	err = os.WriteFile(db.path, marshaledChirp, 0666)
	if err != nil {
		return chirp, err
	}

	return chirp, nil
}

func (db *DB) LoadDB() (*DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	data, err := os.ReadFile(db.path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, db.databaseStructure)
	if err != nil {
		return nil, err
	}

	return db.databaseStructure, nil
}
