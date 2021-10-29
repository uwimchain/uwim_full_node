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
	//TransactionsMemory   = &storage.TransactionsMemory
	NewTx                  = deep_actions.NewTx
	NewComment             = deep_actions.NewComment
	GetDelegateScBalance   = storage.GetBalance(config.DelegateScAddress)
	NewBalance             = deep_actions.NewBalance
	SendTx                 = sender.SendTx
	DonateStandardCardData = deep_actions.DonateStandardCardData{}
	NewBuyTokenSign        = deep_actions.NewBuyTokenSign
	GetBalanceForToken     = storage.GetBalanceForToken
	GetBalance             = storage.GetBalance
	AddTokenEmission       = storage.AddTokenEmission
	GetToken               = deep_actions.GetToken
	GetAddress             = deep_actions.GetAddress
)

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
	//configDataJson, _ := json.Marshal(c.ConfigData)
	//_ = json.Unmarshal(configDataJson, &configData)
	_ = json.Unmarshal([]byte(apparel.ConvertInterfaceToString(c.ConfigData)), &configData)

	return configData
}

type Balance deep_actions.Balance
type BusinessStandardCardData deep_actions.BusinessStandardCardData
type Tx deep_actions.Tx

type ContractCommentData struct {
	NodeAddress string `json:"node_address"`
	CheckSum    []byte `json:"check_sum"`
}

func NewContractCommentData(nodeAddress string, checkSum []byte) *ContractCommentData {
	return &ContractCommentData{
		NodeAddress: nodeAddress,
		CheckSum:    checkSum,
	}
}

func GetTokenInfoForScAddress(scAddress string) *deep_actions.Token {

	publicKey, _ := crypt.PublicKeyFromAddress(scAddress)
	uwAddress := crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey)
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

	txCommentSign := NewBuyTokenSign(
		config.NodeNdAddress,
	)

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	SendNewScTx(timestamp, config.BlockHeight, scAddress, uwAddress, amount, tokenLabel, "refund_transaction", txCommentSign)

	return nil
}

func SendNewScTx(timestampD string, height int64, from, to string, amount float64, tokenLabel, commentTitle string,
	commentData interface{}) {
	commentDataJson, _ := json.Marshal(commentData)

	tx := deep_actions.Tx{
		Type:       5,
		Nonce:      apparel.GetNonce(timestampD),
		HashTx:     "",
		Height:     height,
		From:       from,
		To:         to,
		Amount:     amount,
		TokenLabel: tokenLabel,
		Timestamp:  timestampD,
		Tax:        0,
		Signature:  nil,
		Comment: deep_actions.Comment{
			Title: commentTitle,
			Data:  commentDataJson,
		},
	}

	tx.SetSignature(config.NodeSecretKey)
	tx.SetHash()

	if memory.IsNodeProposer() {
		sender.SendTx(tx)
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	}
}
