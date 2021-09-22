package my_token_con

import (
	"node/blockchain/contracts"
)

var (
	db = contracts.Database{}

	PoolDB = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_pool")
	//TxDB    = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_tx")
	//TxsDB   = db.NewConnection("blockchain/contracts/donate_token_con/storage/donate_token_contract_txs")
	//LogDB   = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_log")
	EventDB  = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_config")
)

type Pool struct {
	Address string `json:"address"`
}