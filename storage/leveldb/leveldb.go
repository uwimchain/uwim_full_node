package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
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

type LevelDb interface {
	NewConnection(path string) *Database
	Put(key string, value string)
	Has(key string) bool
	Get(key string) Row
	GetAll(prefix string) []Row
}

var (
	db        = Database{}
	ConfigDB  = db.NewConnection("storage/config")
	AddressDB = db.NewConnection("storage/address")
	ChainDB   = db.NewConnection("storage/chain")
	TxDB      = db.NewConnection("storage/tx")
	TxsDB     = db.NewConnection("storage/txs")
	TokenDb   = db.NewConnection("storage/token")
)

func (d *Database) NewConnection(path string) *Database {
	con, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatal("Database error:", path, " ", err)
	}

	db := &Database{
		dbConn: con,
	}

	return db
}

func (d Database) Put(key string, value string) {
	err := d.dbConn.Put([]byte(key), []byte(value), nil)
	if err != nil {
		log.Println("Put error:", err)
	}
}

func (d Database) Has(key string) bool {
	result, err := d.dbConn.Has([]byte(key), nil)
	if err != nil {
		log.Println("Has error:", err)
	}

	return result
}

func (d Database) Get(key string) Row {
	result := *NewRow(key, "")

	data, err := d.dbConn.Get([]byte(key), nil)
	if err == nil {
		result.Value = string(data)
	}

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

	err := iter.Error()
	if err != nil {
		log.Println("Get All error:", err)
	}

	iter.Release()

	return rows
}
