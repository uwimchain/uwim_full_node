package api

import (
	"bytes"
	"encoding/json"
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
	"strings"
)

// RenameToken method arguments
type RenameTokenArgs struct {
	Mnemonic string `json:"mnemonic"`
	Proposer string `json:"proposer"`
	Label    string `json:"label"`
	NewName  string `json:"new_name"`
}

func (api *Api) RenameToken(args *RenameTokenArgs, result *string) error {
	args.Mnemonic, args.Proposer, args.Label = apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.Proposer), apparel.TrimToLower(args.Label)

	if check := validateRenameToken(args); check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	token := deep_actions.Token{
		Type:  1,
		Label: args.Label,
		Name:  args.NewName,
	}

	jsonString, _ := json.Marshal(token)
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	comment := deep_actions.Comment{
		Title: "rename_token_transaction",
		Data:  jsonString,
	}

	tx := deep_actions.Tx{
		Type:       3,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       args.Proposer,
		To:         config.NodeNdAddress,
		Amount:     config.RenameTokenCost,
		TokenLabel: "uwm",
		Timestamp:  timestamp,
		Tax:        0,
		Signature:  nil,
		Comment:    comment,
	}

	jsonString, _ = json.Marshal(deep_actions.Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})
	tx.Signature = crypt.SignMessageWithSecretKey(secretKey, jsonString)

	jsonString, _ = json.Marshal(tx)
	tx.HashTx = crypt.GetHash(jsonString)

	sender.SendTx(tx)
	storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	*result = "Token renamed"

	return nil
}

func validateRenameToken(args *RenameTokenArgs) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	if args.Mnemonic == "" {
		return 2
	}

	if len(bytes.Split([]byte(strings.TrimSpace(args.Mnemonic)), []byte(" "))) != 24 {
		return 2
	}

	if args.Proposer == "" {
		return 3
	}

	if !strings.HasPrefix(args.Proposer, metrics.AddressPrefix) {
		return 3
	}

	if len(args.Proposer) != 61 {
		return 3
	}

	if args.Proposer == config.GenesisAddress {
		return 3
	}

	publicKeyFromAddress, _ := crypt.PublicKeyFromAddress(args.Proposer)
	publicKeyFromMnemonic := crypt.PublicKeyFromSecretKey(crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)))
	if string(publicKeyFromAddress) != string(publicKeyFromMnemonic) {
		return 4
	}

	if args.Label == "" {
		return 5
	}

	if !storage.CheckToken(args.Label) {
		return 6
	}

	if args.NewName == "" {
		return 7
	}

	if int64(len(args.NewName)) > config.MaxName {
		return 8
	}

	balance := storage.GetBalance(args.Proposer)
	if balance != nil {
		for _, coin := range balance {
			if coin.TokenLabel == "uwm" {
				if coin.Amount < config.RenameTokenCost {
					return 9
				}
			}
		}
	} else {
		return 9
	}

	token := storage.GetAddressToken(args.Proposer)
	if token.Label == "" {
		return 10
	}

	if args.Label == "uwm" {
		return 11
	}

	return 0
}
