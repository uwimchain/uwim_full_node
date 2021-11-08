package business_token_con

import (
	"encoding/json"
	"node/blockchain/contracts"
)

var (
	db = contracts.Database{}

	ContractsDB = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_contracts")
	EventDB     = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_event")
	ConfigDB    = db.NewConnection("blockchain/contracts/business_token_con/storage/business_token_contract_config")
)

type Partner struct {
	Address string              `json:"address"`
	Percent float64             `json:"percent"`
	Balance []contracts.Balance `json:"balance"`
}

type Partners []Partner

func GetPartners(scAddress string) Partners {
	partnersJson := ContractsDB.Get(scAddress).Value
	if partnersJson != "" {
		partners := Partners{}
		_ = json.Unmarshal([]byte(partnersJson), &partners)
		return partners
	}

	return nil
}

func (ps *Partners) Update(scAddress string) {
	jsonString, _ := json.Marshal(ps)

	ContractsDB.Put(scAddress, string(jsonString))
}

func ClearPartners(scAddress string) {
	ContractsDB.Put(scAddress, "")
}
