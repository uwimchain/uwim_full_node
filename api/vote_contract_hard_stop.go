package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts/vote_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

// VoteContractHardStop method arguments
type VoteContractHardStopArgs struct {
	Mnemonic  string `json:"mnemonic"`
	VoteNonce string `json:"vote_nonce"`
}

func (api *Api) VoteContractHardStop(args *VoteContractHardStopArgs, result *string) error {
	args.Mnemonic = apparel.TrimToLower(args.Mnemonic)

	starterAddress := crypt.AddressFromMnemonic(args.Mnemonic)
	validateStart := validateHardStop(args.Mnemonic, starterAddress, args.VoteNonce)
	if validateStart != 0 {
		return errors.New(strconv.FormatInt(validateStart, 10))
	}

	commentData := make(map[string]interface{})
	commentData["vote_nonce"] = args.VoteNonce

	commentDataJson, err := json.Marshal(commentData)
	if err != nil {
		log.Println("Send Transaction error 1:", err)
	}

	timestamp := apparel.TimestampUnix()

	comment := deep_actions.Comment{
		Title: "vote_contract_hard_stop_transaction",
		Data:  commentDataJson,
	}

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(strconv.FormatInt(timestamp, 10)),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       starterAddress,
		To:         config.VoteScAddress,
		Amount:     0,
		TokenLabel: config.BaseToken,
		Timestamp:  strconv.FormatInt(timestamp, 10),
		Tax:        0,
		Signature:  nil,
		Comment:    comment,
	}

	jsonString, _ := json.Marshal(deep_actions.Tx{
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

	if memory.IsValidator() {
		storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	}

	*result = "Transaction send"

	return nil
}

func validateHardStop(mnemonic, starterAddress string, voteNonce string) int64 {
	validateMnemonic := validateMnemonic(mnemonic, starterAddress)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateHardStop := vote_con.ValidateHardStop(starterAddress, voteNonce)
	if validateHardStop != 0 {
		return validateHardStop
	}

	validateTxInMemory := validateTxInMemory(starterAddress, config.VoteSuperAddress, "vote_contract_hard_stop_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}
	return 0
}
