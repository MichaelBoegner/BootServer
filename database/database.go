package database

import (
	"encoding/json"
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
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := os.ReadFile("database/database.json")
	fmt.Printf("THIS IS ERR == %v, THIS IS DATA == %v", err, data)

	databaseStructure := DBStructure{
		Chirps: make(map[int]Chirp),
	}

	if err != nil {
		marshaledData, err := json.Marshal(databaseStructure)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(db.path, []byte(marshaledData), 0666)
		return db, err
	} else {
		marshaledData, err := json.Marshal(databaseStructure)
		if err != nil {
			return nil, err
		}
		os.Remove("database/database.json")
		err = os.WriteFile(db.path, []byte(marshaledData), 0666)
		return db, err
	}
}

func (db *DB) LoadDB() (*DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := os.ReadFile(db.path)

	databaseStructure := &DBStructure{}
	err = json.Unmarshal(data, databaseStructure)

	if err != nil {
		return nil, err
	}

	return databaseStructure, nil
}
