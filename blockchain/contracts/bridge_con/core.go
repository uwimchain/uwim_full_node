package bridge_con

import "node/blockchain/contracts"

var (
	db = contracts.Database{}

	TxsDB    = db.NewConnection("blockchain/contracts/bridge_con/storage/bridge_contract_txs")
	EventDB  = db.NewConnection("blockchain/contracts/bridge_con/storage/bridge_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/bridge_con/storage/bridge_contract_config")
)

type MemoryTx struct {
	Hash string `json:"hash"`
	// system types
	// 0 - ethereum
	// 1 - binance
	System        int     `json:"system"`
	Amount        float64 `json:"amount"`
	Timestamp     string  `json:"timestamp"`
	Address       string  `json:"address"`
	SystemAddress string  `json:"system_address"`
	// statuses
	// 0 - processing
	// 1 - success
	// 2 - deny
	Status int `json:"status"`
}

func NewMemoryTx(hash string, system int, amount float64, timestamp string, address string, systemAddress string, status int) *MemoryTx {
	return &MemoryTx{Hash: hash, System: system, Amount: amount, Timestamp: timestamp, Address: address, SystemAddress: systemAddress, Status: status}
}

func (mTx *MemoryTx) AddToMemory() {
	Memory = append(Memory, *mTx)
}

var Memory []MemoryTx

// renew trial scorpion proud observe hard security twin media parent tiger alcohol tourist other rack million lizard version spread next marble club else recycle
var ScAddress string = ""

// blue quick expand train vehicle north loan tissue grief exhaust stock table quarter awful brain club fringe purchase drift explain room start mouse injury
var AdminUwAddress string = ""