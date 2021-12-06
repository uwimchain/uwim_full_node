package trade_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
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
	amount = apparel.Round(amount)
	return &TradeArgs{ScAddress: scAddress, UwAddress: uwAddress, Amount: amount, TokenLabel: tokenLabel, BlockHeight: blockHeight, TxHash: txHash}
}

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
	FirstToken  PoolToken `json:"first_token"`
	SecondToken PoolToken `json:"second_token"`
	Liq         Liq       `json:"liq"`
}

type Liq struct {
	Amount     float64          `json:"amount"`
	UpdateTime contracts.String `json:"update_time"`
}

type PoolToken struct {
	Amount     float64          `json:"amount"`
	UpdateTime contracts.String `json:"update_time"`
	Commission float64          `json:"commission"`
}

type Holder struct {
	Address string `json:"address"`
	Pool    Pool   `json:"pool"`
}

type Holders []Holder

func AddToken(scAddress string) error {
	var scAddressHolders Holders

	scAddressConfig := contracts.Config{
		LastEventHash: "",
		ConfigData: TradeConfig{
			Commission: 0,
		},
	}

	scAddressPool := Pool{
		FirstToken: PoolToken{
			Amount:     0,
			UpdateTime: "",
			Commission: 0,
		},
		SecondToken: PoolToken{
			Amount:     0,
			UpdateTime: "",
			Commission: 0,
		},
		Liq: Liq{
			Amount:     0,
			UpdateTime: "",
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

func GetHolders() map[string]Holders {
	holdersRows := HolderDB.GetAll("")
	tokenHolders := make(map[string]Holders)

	if holdersRows != nil {
		for _, i := range holdersRows {
			var holder Holder
			_ = json.Unmarshal([]byte(i.Value), &holder)
			tokenHolders[i.Key] = append(tokenHolders[i.Key], holder)
		}
	}

	return tokenHolders
}

func FixAmount(recipient string, amount float64, tokenLabel string) float64 {
	if !crypt.IsAddressSmartContract(recipient) {
		return 0
	}

	poolJson := PoolDB.Get(recipient).Value
	pool := Pool{}
	_ = json.Unmarshal([]byte(poolJson), &pool)

	var course float64 = 0
	if pool.SecondToken.Amount > 0 && pool.FirstToken.Amount > 0 {
		course = pool.FirstToken.Amount / pool.SecondToken.Amount
	}

	token := contracts.GetTokenInfoForScAddress(recipient)
	if token.Label == "" {
		return 0
	}

	switch tokenLabel {
	case config.BaseToken:
		amount = amount / course
		break
	case token.Label:
		amount = amount * course
		break
	}

	return amount
}
