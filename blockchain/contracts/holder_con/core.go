package holder_con

import (
	"encoding/json"
	"node/blockchain/contracts"
)

var (
	db = contracts.Database{}

	HolderDB = db.NewConnection("blockchain/contracts/holder_con/storage/holder_contract_holder")
	EventDB  = db.NewConnection("blockchain/contracts/holder_con/storage/holder_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/holder_con/storage/holder_contract_config")
)

type Holder struct {
	DepositorAddress string  `json:"depositor_address"`
	RecipientAddress string  `json:"recipient_address"`
	Amount           float64 `json:"amount"`
	TokenLabel       string  `json:"token_label"`
	GetBlockHeight   int64   `json:"get_block_height"`
}

type Holders []Holder

func (hs *Holders) Update(address string) {
	jsonString, _ := json.Marshal(hs)

	HolderDB.Put(address, string(jsonString))
}
