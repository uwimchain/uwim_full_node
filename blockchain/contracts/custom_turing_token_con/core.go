package custom_turing_token_con

import "node/blockchain/contracts"

var (
	db = contracts.Database{}

	HolderDB = db.NewConnection("blockchain/contracts/custom_turing_token_con/storage/custom_turing_contract_holder")
	EventDB  = db.NewConnection("blockchain/contracts/custom_turing_token_con/storage/custom_turing_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/custom_turing_token_con/storage/custom_turing_contract_config")
)

type Holder struct {
	Address    string  `json:"address"`
	Amount     float64 `json:"amount"`
	UpdateTime string  `json:"update_time"`
}

var (
	ScAddress        string  = "sc1jug0957xjgjef09utda75gkhfsxphcjw3gq8nh4sx6hq6a65v64sqazr60"
	UwAddress        string  = "uw1jug0957xjgjef09utda75gkhfsxphcjw3gq8nh4sx6hq6a65v64saku3rn"
	ScAddressPercent float64 = 15
	TokenLabel       string  = "artx"
)
