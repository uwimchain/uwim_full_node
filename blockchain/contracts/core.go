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
	TransactionsMemory     = &storage.TransactionsMemory
	NewTx                  = deep_actions.NewTx
	NewComment             = deep_actions.NewComment
	StorageA               = deep_actions.Address{}
	StorageUpdateBalance   = StorageA.UpdateBalance
	GetDelegateScBalance   = storage.GetBalance(config.DelegateScAddress)
	NewBalance             = deep_actions.NewBalance
	SendTx                 = sender.SendTx
	GetAddress             = storage.GetAddress
	CheckAddressToken      = storage.CheckAddressToken
	GetAddressToken        = storage.GetAddressToken
	DonateStandardCardData = deep_actions.DonateStandardCardData{}
	//BusinessStandardCardData = deep_actions.BusinessStandardCardData{}
	NewBuyTokenSign    = deep_actions.NewBuyTokenSign
	GetBalanceForToken = storage.GetBalanceForToken
	GetBalance         = storage.GetBalance
)

type Config struct {
	/*Commission    float64 `json:"commission"`*/
	LastEventHash string      `json:"last_event_hash"`
	ConfigData    interface{} `json:"config_data"`
}

type Balance deep_actions.Balance
type BusinessStandardCardData deep_actions.BusinessStandardCardData
type Tx deep_actions.Tx
type Log struct {
	Timestamp  int64  `json:"timestamp"`
	TimestampD string `json:"timestamp_d"`
	Record     string `json:"record"`
}

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

func GetTokenInfoForScAddress(scAddress string) deep_actions.Token {
	token := deep_actions.Token{}

	publicKey, _ := crypt.PublicKeyFromAddress(scAddress)
	if publicKey != nil {
		uwAddress := crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey)
		token = GetAddressToken(uwAddress)
	}

	return token
}

// function for refund user token pairs
func RefundTransaction(scAddress string, uwAddress string, amount float64, tokenLabel string) error { // test
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

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
	tx := NewTx(
		5,
		apparel.GetNonce(timestamp),
		"",
		config.BlockHeight,
		scAddress,
		uwAddress,
		amount,
		tokenLabel,
		timestamp,
		0,
		nil,
		*NewComment("refund_transaction", nil),
	)

	jsonString, _ := json.Marshal(Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})
	tx.Signature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)

	SendTx(*tx)
	*TransactionsMemory = append(*TransactionsMemory, *tx)

	return nil
}