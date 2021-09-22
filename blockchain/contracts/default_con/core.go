package default_con

import "node/blockchain/contracts"

var (
	db = contracts.Database{}

	EventDB  = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_config")
	TokenDB  = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_token")
)

type NftToken struct {
	Id        int64   `json:"id"`        // уникальный id
	Hash      []byte  `json:"hash"`      // хэш TODO: подобрать шифрование
	Owner     string  `json:"owner"`     // адрес смарт-контракта
	Price     float64 `json:"price"`     // округление до 12 символов после запятой
	Timestamp string  `json:"timestamp"` // в строке timestamp unix
}
