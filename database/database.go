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
		err := os.WriteFile(db.path, []byte("My first Chrip."), 0666)
		return db, err
	} else {
		os.Remove("database/database.json")
		err := os.WriteFile(db.path, []byte("My first Chrip."), 0666)
		return db, err
	}
}

func (db *DB) LoadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := os.ReadFile(db.path)

	databaseStructre := DBStructure{
		Chirps: make(map[int]Chirp),
	}
	databaseStructre.Chirps[0] = Chirp{Body: string(data)}

	if err != nil {
		return databaseStructre, err
	}

	return databaseStructre, nil
}
