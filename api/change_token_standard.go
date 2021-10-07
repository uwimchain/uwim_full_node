package api

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
	"strconv"
)

// ChangeTokenStandard method arguments
type ChangeTokenStandardArgs struct {
	Mnemonic string `json:"mnemonic"`
	// 0 - My
	// 1 - Donate
	// 3 - StartUp
	// 4 - Business
	// 5 - Trade
	Standard int64 `json:"standard"`
}

func (api *Api) ChangeTokenStandard(args *ChangeTokenStandardArgs, result *string) error {

	args.Mnemonic = apparel.TrimToLower(args.Mnemonic)

	proposer := crypt.AddressFromMnemonic(args.Mnemonic)

	if check := validateChangeTokenStandard(args.Mnemonic, proposer, args.Standard); check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	address := deep_actions.GetAddress(proposer)
	t := deep_actions.GetToken(address.TokenLabel)

	if t.Label == "" {
		return errors.New(strconv.FormatInt(9, 10))
	}

	token := deep_actions.Token{
		Label:    t.Label,
		Standard: args.Standard,
	}

	commentData, _ := json.Marshal(token)
	comment := deep_actions.Comment{
		Title: "change_token_standard_transaction",
		Data:  commentData,
	}

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	tx := deep_actions.Tx{
		Type:       3,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       proposer,
		To:         config.NodeNdAddress,
		Amount:     config.ChangeTokenStandardCost,
		TokenLabel: "uwm",
		Timestamp:  timestamp,
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
	storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	*result = "Token standard changed"

	return nil
}

func validateChangeTokenStandard(mnemonic, proposer string, standard int64) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	if check := validateMnemonic(mnemonic, proposer); check != 0 {
		return check
	}

	if !apparel.SearchInArray([]int64{1, 3, 4, 5}, standard) {
		return 5
	}

	address := deep_actions.GetAddress(proposer)
	token := deep_actions.GetToken(address.TokenLabel)
	if token == nil {
		return 7
	}

	if standard == token.Standard {
		return 8
	}

	if token.Standard == 0 && !apparel.SearchInArray([]int64{1, 3, 4, 5}, standard) {
		return 9
	}

	if token.Standard == 1 && !apparel.SearchInArray([]int64{3, 4, 5}, standard) {
		return 9
	}

	if token.Standard == 3 && !apparel.SearchInArray([]int64{4, 6}, standard) {
		return 9
	}

	if token.Standard == 7 || token.Standard == 2 {
		return 9
	}

	if check := validateBalance(proposer, config.ChangeTokenStandardCost, config.BaseToken, true); check != 0 {
		return check
	}

	return 0
}
