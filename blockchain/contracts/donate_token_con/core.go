package donate_token_con

import (
	"node/blockchain/contracts"
)

var (
	db = contracts.Database{}

	EventDB  = db.NewConnection("blockchain/contracts/donate_token_con/storage/donate_token_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/donate_token_con/storage/donate_token_contract_config")
)