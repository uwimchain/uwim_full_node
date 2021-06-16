package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/crypt"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"strconv"
)

// Balance method arguments
type BalanceArgs struct {
	Address string `json:"address"`
}

type Info struct {
	Address       BalanceInfo `json:"address"`
	SmartContract BalanceInfo `json:"smartContract"`
	Node          BalanceInfo `json:"node"`
}

type BalanceInfo struct {
	Address      string                 `json:"address"`
	Balance      []deep_actions.Balance `json:"balance"`
	Transactions []deep_actions.Tx      `json:"transactions"`
	Token        deep_actions.Token     `json:"token"`
	ScKeeping    bool                   `json:"sc_keeping"`
}

func (api *Api) Balance(args *BalanceArgs, result *string) error {
	if args.Address != "" && args.Address != "0" {
		publicKey, err := crypt.PublicKeyFromAddress(args.Address)
		if err != nil {
			log.Println("Api Balance error 1:", err)
		}
		jsonString, err := json.Marshal(Info{
			Address: BalanceInfo{
				Address:      crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey),
				Balance:      storage.GetBalance(crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey)),
				Transactions: storage.GetTransactions(crypt.AddressFromPublicKey(metrics.AddressPrefix, publicKey)),
				Token:        storage.GetAddressToken(args.Address),
				ScKeeping:    storage.CheckAddressScKeeping(args.Address),
			},
			SmartContract: BalanceInfo{
				Address:      crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey),
				Balance:      storage.GetBalance(crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)),
				Transactions: storage.GetTransactions(crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)),
				Token:        storage.GetAddressToken(args.Address),
				ScKeeping:    storage.CheckAddressScKeeping(args.Address),
			},
			Node: BalanceInfo{
				Address:      crypt.AddressFromPublicKey(metrics.NodePrefix, publicKey),
				Balance:      storage.GetBalance(crypt.AddressFromPublicKey(metrics.NodePrefix, publicKey)),
				Transactions: storage.GetTransactions(crypt.AddressFromPublicKey(metrics.NodePrefix, publicKey)),
				Token:        storage.GetAddressToken(args.Address),
				ScKeeping:    storage.CheckAddressScKeeping(args.Address),
			},
		})
		if err != nil {
			log.Println("Api Balance error 2:", err)
		}
		*result = string(jsonString)
		return nil
	} else {
		return errors.New(strconv.Itoa(1))
	}
}
