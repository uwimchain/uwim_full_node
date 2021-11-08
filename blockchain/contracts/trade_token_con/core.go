package trade_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
)

var (
	db = contracts.Database{}

	PoolDB   = db.NewConnection("blockchain/contracts/trade_token_con/storage/trade_token_contract_pool")
	HolderDB = db.NewConnection("blockchain/contracts/trade_token_con/storage/trade_token_contract_holder")
	EventDB  = db.NewConnection("blockchain/contracts/trade_token_con/storage/trade_token_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/trade_token_con/storage/trade_token_contract_config")
)

type TradeArgs struct {
	ScAddress   string  `json:"sc_address"`
	UwAddress   string  `json:"address"`
	Amount      float64 `json:"amount"`
	TokenLabel  string  `json:"token_label"`
	BlockHeight int64   `json:"block_height"`
	TxHash      string  `json:"tx_hash"`
}

func NewTradeArgs(scAddress string, uwAddress string, amount float64, tokenLabel string, blockHeight int64, txHash string) *TradeArgs {
	//amount, _ = apparel.Round(amount)
	amount = apparel.Round(amount)
	return &TradeArgs{ScAddress: scAddress, UwAddress: uwAddress, Amount: amount, TokenLabel: tokenLabel, BlockHeight: blockHeight, TxHash: txHash}
}

func NewTradeArgsForValidate(scAddress string, uwAddress string, amount float64, tokenLabel string) *TradeArgs {
	//amount, _ = apparel.Round(amount)
	amount = apparel.Round(amount)
	return &TradeArgs{ScAddress: scAddress, UwAddress: uwAddress, Amount: amount, TokenLabel: tokenLabel}
}

/*func NewTradeArgs(scAddress string, uwAddress string, amount float64, tokenLabel string) *TradeArgs {
	return &TradeArgs{ScAddress: scAddress, UwAddress: uwAddress, Amount: amount, TokenLabel: tokenLabel}
}*/

type GetArgs struct {
	ScAddress   string `json:"sc_address"`
	UwAddress   string `json:"uw_address"`
	TokenLabel  string `json:"token_label"`
	BlockHeight int64  `json:"block_height"`
	TxHash      string `json:"tx_hash"`
}

func NewGetArgs(scAddress string, uwAddress string, tokenLabel string, blockHeight int64, txHash string) *GetArgs {
	return &GetArgs{ScAddress: scAddress, UwAddress: uwAddress, TokenLabel: tokenLabel, BlockHeight: blockHeight, TxHash: txHash}
}

type TradeConfig struct {
	Commission float64 `json:"commission"`
}

type Pool struct {
	FirstToken  PoolToken `json:"first_token"`  // uwm
	SecondToken PoolToken `json:"second_token"` // user token
	Liq         Liq       `json:"liq"`
}

type Liq struct {
	Amount     float64 `json:"amount"`
	UpdateTime int64   `json:"update_time"`
}

type PoolToken struct {
	Amount     float64 `json:"amount"`
	UpdateTime int64   `json:"update_time"`
	Commission float64 `json:"commission"`
}

type Holder struct {
	Address string `json:"address"`
	Pool    Pool   `json:"pool"`
}

func AddToken(scAddress string) error {
	var scAddressHolders []Holder

	scAddressConfig := contracts.Config{
		LastEventHash: "",
		ConfigData: TradeConfig{
			Commission: 0,
		},
	}

	scAddressPool := Pool{
		FirstToken: PoolToken{
			Amount:     0,
			UpdateTime: 0,
			Commission: 0,
		},
		SecondToken: PoolToken{
			Amount:     0,
			UpdateTime: 0,
			Commission: 0,
		},
		Liq: Liq{
			Amount:     0,
			UpdateTime: 0,
		},
	}

	jsonScAddressPool, err := json.Marshal(scAddressPool)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: %v", err))
	}

	jsonScAddressHolders, err := json.Marshal(scAddressHolders)
	if err != nil {
		return errors.New(fmt.Sprintf("error 2: %v", err))
	}

	jsonScAddressConfig, err := json.Marshal(scAddressConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("error 3: %v", err))
	}

	PoolDB.Put(scAddress, string(jsonScAddressPool))
	HolderDB.Put(scAddress, string(jsonScAddressHolders))
	ConfigDB.Put(scAddress, string(jsonScAddressConfig))
	return nil
}
