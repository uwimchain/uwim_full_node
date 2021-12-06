package contracts

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

var (
	GetBalanceForToken = storage.GetBalanceForToken
	GetBalance         = storage.GetBalance
	AddTokenEmission   = storage.AddTokenEmission
	GetToken           = deep_actions.GetToken
	GetAddress         = deep_actions.GetAddress
)

type String string

func (s *String) UnmarshalJSON(b []byte) error {
	var item interface{}
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}

	switch v := item.(type) {
	case int64:
		*s = String(strconv.FormatInt(v, 10))
		break
	case float64:
		*s = String(strconv.Itoa(int(v)))
		break
	case string:
		*s = String(v)
	}
	return nil
}

type Config struct {
	LastEventHash string      `json:"last_event_hash"`
	ConfigData    interface{} `json:"config_data"`
}

func GetConfig(configDb *Database, scAddress string) *Config {
	configJson := configDb.Get(scAddress).Value
	configObj := Config{}

	_ = json.Unmarshal([]byte(configJson), &configObj)

	return &configObj
}

func (c *Config) GetData() map[string]interface{} {
	configData := make(map[string]interface{})

	if c.ConfigData != nil {
		jsonString, _ := json.Marshal(c.ConfigData)
		_ = json.Unmarshal(jsonString, &configData)
	}

	return configData
}

func (c *Config) Update(configDb *Database, scAddress string) {
	jsonString, _ := json.Marshal(c)

	configDb.Put(scAddress, string(jsonString))
}

type Balance deep_actions.Balance
type Tx deep_actions.Tx
type Txs deep_actions.Txs

func GetTokenInfoForScAddress(scAddress string) *deep_actions.Token {
	uwAddress := crypt.AddressFromAnotherAddress(metrics.AddressPrefix, scAddress)
	address := GetAddress(uwAddress)
	return address.GetToken()
}

func RefundTransaction(scAddress string, uwAddress string, amount float64, tokenLabel string) error {
	if !memory.IsNodeProposer() {
		return nil
	}

	scBalance := GetBalance(scAddress)
	if scBalance == nil {
		return errors.New("error 1: sc balance is null")
	}

	check := false
	for _, i := range scBalance {
		if i.TokenLabel == tokenLabel {
			if i.Amount < amount {
				return errors.New(fmt.Sprintf("error 2: smart contract has low balance for token %s. Has %g, but need %g", tokenLabel, i.Amount, amount))
			}

			check = true
			break
		}
	}

	if !check {
		return errors.New(fmt.Sprintf("error 3: samrt contract balance haven`t token %s", tokenLabel))
	}

	SendNewScTx(scAddress, uwAddress, amount, tokenLabel, "refund_transaction")

	return nil
}

func SendNewScTx(from, to string, amount float64, tokenLabel, commentTitle string) {
	sign, _ := json.Marshal(*deep_actions.NewBuyTokenSign(config.NodeNdAddress))

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	tx := deep_actions.Tx{
		Type:       5,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       from,
		To:         to,
		Amount:     amount,
		TokenLabel: tokenLabel,
		Timestamp:  timestamp,
		Tax:        0,
		Signature:  nil,
		Comment: deep_actions.Comment{
			Title: commentTitle,
			Data:  sign,
		},
	}

	tx.SetSignature(config.NodeSecretKey)
	tx.SetHash()

	if memory.IsNodeProposer() {
		sender.SendTx(tx)
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	}
}
