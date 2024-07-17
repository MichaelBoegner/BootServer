package database

import (
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{}
	db.path = "./database.json"
	db.mux = &sync.RWMutex{}
	_, err := os.ReadFile(db.path)
	if err != nil {
		os.WriteFile(db.path, []byte(`{"chirps":{}}`), 0644)
		return db, err
	}
	return db, nil
}
