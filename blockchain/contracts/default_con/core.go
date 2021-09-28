package default_con

import "node/blockchain/contracts"

var (
	db = contracts.Database{}

	EventDB  = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_config")
	TokenDB  = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_token")
)

type NftToken struct {
	Id        int64   `json:"id"`
	Hash      []byte  `json:"hash"`
	Owner     string  `json:"owner"`
	Price     float64 `json:"price"`
	Timestamp string  `json:"timestamp"`
}
