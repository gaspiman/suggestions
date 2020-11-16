package datastore

import "github.com/dgraph-io/badger/v2"

const DB_PATH = "/data/db"

type Datastore struct {
	DB *badger.DB
}

func NewDatastore() (*Datastore, error) {
	ds := &Datastore{}
	db, err := badger.Open(badger.DefaultOptions(DB_PATH))
	if err != nil {
		return nil, err
	}
	ds.DB = db
	return ds, nil
}
