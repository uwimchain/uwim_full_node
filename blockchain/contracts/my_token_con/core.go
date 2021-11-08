package my_token_con

import (
	"node/blockchain/contracts"
)

var (
	db = contracts.Database{}

	PoolDB = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_pool")
	EventDB  = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_config")
)

type Pool struct {
	Address string `json:"address"`
}