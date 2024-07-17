package database

import (
	"fmt"
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
	db.path = path
	db.mux = &sync.RWMutex{}
	data, err := os.ReadFile("database/database.json")
	fmt.Printf("THIS IS ERR == %v, THIS IS DATA == %v", err, data)
	if err != nil {
		err := os.WriteFile(db.path, []byte("{'test'}"), 0666)
		return db, err
	}
	return db, nil
}
