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

// FillTokenCard method arguments
type FillTokenCardArgs struct {
	Mnemonic   string                   `json:"mnemonic"`
	Proposer   string                   `json:"proposer"`
	FullName   string                   `json:"full_name"`
	BirthDay   string                   `json:"birthday"`
	Gender     string                   `json:"gender"`
	Country    string                   `json:"country"`
	Region     string                   `json:"region"`
	City       string                   `json:"city"`
	Social     *deep_actions.Social     `json:"social"`
	Messengers *deep_actions.Messengers `json:"messengers"`
	Email      string                   `json:"email"`
	Site       string                   `json:"site"`
	Hashtags   string                   `json:"hashtags"`
	WorkPlace  string                   `json:"work_place"`
	School     string                   `json:"school"`
	Education  string                   `json:"education"`
	Comment    string                   `json:"comment"`
}

func (api *Api) FillTokenCard(args *FillTokenCardArgs, result *string) error {
	args.Mnemonic, args.Proposer, args.FullName, args.BirthDay, args.Gender, args.Country, args.Region, args.City,
		args.Email, args.Site, args.Hashtags, args.WorkPlace, args.School, args.Education, args.Comment =
		apparel.TrimToLower(args.Mnemonic), apparel.TrimToLower(args.Proposer), strings.TrimSpace(args.FullName),
		strings.TrimSpace(args.BirthDay), strings.TrimSpace(args.Gender), strings.TrimSpace(args.Country),
		strings.TrimSpace(args.Region), strings.TrimSpace(args.City), strings.TrimSpace(args.Email),
		strings.TrimSpace(args.Site), strings.TrimSpace(args.Hashtags), strings.TrimSpace(args.WorkPlace),
		strings.TrimSpace(args.School), strings.TrimSpace(args.Education), strings.TrimSpace(args.Comment)

	if check := validateCardFields(args); check != 0 {
		return errors.New(strconv.FormatInt(check, 10))
	}

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	tokenCard := deep_actions.PersonalTokenCard{
		FullName:   args.FullName,
		BirthDay:   args.BirthDay,
		Gender:     args.Gender,
		Country:    args.Country,
		Region:     args.Region,
		City:       args.City,
		Social:     args.Social,
		Messengers: args.Messengers,
		Email:     args.Email,
		Site:      args.Site,
		Hashtags:  args.Hashtags,
		WorkPlace: args.WorkPlace,
		School:    args.School,
		Education: args.Education,
		Comment:   args.Comment,
	}

	jsonString, _ := json.Marshal(tokenCard)
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	comment := deep_actions.Comment{
		Title: "fill_token_card_transaction",
		Data:  jsonString,
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
	*result = "Token card filled"
	return nil
}

func validateCardFields(args *FillTokenCardArgs) int64 {
	if !memory.IsMainNode() {
		return 1
	}

	if check := validateMnemonic(args.Mnemonic, args.Proposer); check != 0 {
		return check
	}

	if check := validateBalance(args.Proposer, config.FillTokenCardCost, config.BaseToken, true); check != 0 {
		return check
	}

	if args.Hashtags != "" {
		if check := validateHashtags(args.Hashtags); check != 0 {
			return check
		}
	}

	return 0
}

func validateHashtags(hashtagsString string) int64 {
	if hashtagsString == "" {
		return 21
	}

	hashtags := strings.Split(strings.TrimSpace(hashtagsString), "#")
	if len(hashtags)-1 < 3 {
		return 22
	}

	if len(hashtags)-1 > 10 {
		return 23
	}

	return 0
}
