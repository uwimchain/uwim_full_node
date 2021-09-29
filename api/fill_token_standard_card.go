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
	"strings"
)

// FillTokenStandardCard method arguments
type FillTokenStandardCardArgs struct {
	Mnemonic             string `json:"mnemonic"`
	Proposer             string `json:"proposer"`
	StandardCardDataJson string `json:"standard_card_data"`
}

func (api *Api) FillTokenStandardCard(args *FillTokenStandardCardArgs, result *string) error {
	args.Mnemonic, args.Proposer, args.StandardCardDataJson =
		apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.Proposer), strings.TrimSpace(args.StandardCardDataJson)

	if check := validateStandardCardFields(args); check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	comment := deep_actions.Comment{
		Title: "fill_token_standard_card_transaction",
		Data:  []byte(args.StandardCardDataJson),
	}

	tx := deep_actions.Tx{
		Type:       3,
		Nonce:      apparel.GetNonce(timestamp),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       args.Proposer,
		To:         config.NodeNdAddress,
		Amount:     config.FillTokenCardCost,
		TokenLabel: config.BaseToken,
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
	*result = "Token standard card filled"

	return nil
}

func validateStandardCardFields(args *FillTokenStandardCardArgs) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	if check := validateMnemonic(args.Mnemonic, args.Proposer); check != 0 {
		return check
	}

	if check := validateBalance(args.Proposer, config.FillTokenStandardCardCost, config.BaseToken, true); check != 0 {
		return check
	}

	if !storage.CheckAddressToken(args.Proposer) {
		return 7
	}

	token := storage.GetAddressToken(args.Proposer)
	switch token.Standard {
	case 2:
		if check := validate2standard(args.StandardCardDataJson); check != 0 {
			return check
		}
		break
	case 3:
		if check := validate3standard(args.StandardCardDataJson); check != 0 {
			return check
		}
		break
	case 4:
		if check := validate4standard(args.StandardCardDataJson); check != 0 {
			return check
		}
		break
	}

	return 0
}

func validate2standard(data string) int64 {
	tokenStandardCard := deep_actions.DonateStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return 121
	}

	return 0
}

func validate3standard(data string) int64 {
	tokenStandardCard := deep_actions.StartUpStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return 131
	}

	return 0
}

func validate4standard(data string) int64 {
	tokenStandardCard := deep_actions.BusinessStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return 141
	}

	if tokenStandardCard.Partners != nil {
		for _, i := range tokenStandardCard.Partners {
			if len(i.Address) != 61 || !crypt.IsAddressUw(i.Address) {
				return 142
			}
		}
	}

	return 0
}
