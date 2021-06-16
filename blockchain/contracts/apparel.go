package contracts

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Database struct {
	dbConn *leveldb.DB
}

type Row struct {
	Key   string
	Value string
}

func NewRow(key string, value string) *Row {
	return &Row{
		Key:   key,
		Value: value,
	}
}

func (d Database) NewConnection(path string) *Database {
	con, _ := leveldb.OpenFile(path, nil)

	db := &Database{
		dbConn: con,
	}

	return db
}

func (d Database) Put(key string, value string) {
	_ = d.dbConn.Put([]byte(key), []byte(value), nil)
}

func (d Database) Has(key string) bool {
	result, _ := d.dbConn.Has([]byte(key), nil)
	return result
}

func (d Database) Get(key string) Row {
	result := *NewRow(key, "")

	data, _ := d.dbConn.Get([]byte(key), nil)
	result.Value = string(data)

	return result
}

func (d Database) GetAll(prefix string) []Row {
	var iter iterator.Iterator
	if prefix == "" {
		iter = d.dbConn.NewIterator(nil, nil)
	} else {
		iter = d.dbConn.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	}

	var rows []Row

	for iter.Next() {
		rows = append(rows, *NewRow(string(iter.Key()), string(iter.Value())))
	}

	_ = iter.Error()

	iter.Release()

	return rows
}
