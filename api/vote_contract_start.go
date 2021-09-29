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

// VoteContractStart method arguments
type VoteContractStartArgs struct {
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	AnswerOptions  interface{} `json:"answer_options"`
	EndBlockHeight int64       `json:"end_block_height"`
	Mnemonic       string      `json:"mnemonic"`
}

func (api *Api) VoteContractStart(args *VoteContractStartArgs, result *string) error {
	args.Mnemonic = apparel.TrimToLower(args.Mnemonic)

	starterAddress := crypt.AddressFromMnemonic(args.Mnemonic)
	validateStart := validateStart(args.Mnemonic, starterAddress, args.Title, args.Description, args.AnswerOptions, args.EndBlockHeight)
	if validateStart != 0 {
		return errors.New(strconv.FormatInt(validateStart, 10))
	}

	answerOptionsJson, err := json.Marshal(args.AnswerOptions)
	if err != nil {
		log.Println("Send Transaction error 1:", err)
	}

	var answerOptions []vote_con.AnswerOption
	if answerOptionsJson != nil {
		err := json.Unmarshal(answerOptionsJson, &answerOptions)
		if err != nil {
			log.Println("Send Transaction error 1:", err)
		}
	}

	if answerOptions != nil {
		for idx := range answerOptions {
			timestamp := apparel.TimestampUnix()

			if answerOptions[idx].AnswerCost <= 0 {
				answerOptions[idx].AnswerCost = config.VoteAnswerOptionDefaultCost
			}

			answerOptions[idx].PossibleAnswerNonce = strconv.FormatInt(apparel.GetNonce(strconv.FormatInt(timestamp, 10)), 10)
		}
	}

	commentData := make(map[string]interface{})
	commentData["title"] = args.Title
	commentData["description"] = args.Description
	commentData["answer_options"] = answerOptions
	commentData["end_block_height"] = args.EndBlockHeight

	commentDataJson, err := json.Marshal(commentData)
	if err != nil {
		log.Println("Send Transaction error 1:", err)
	}

	timestamp := apparel.TimestampUnix()

	comment := deep_actions.Comment{
		Title: "vote_contract_start_transaction",
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
		Signature:  crypt.SignMessageWithSecretKey(crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic)), []byte(starterAddress)),
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

func validateStart(mnemonic, starterAddress, title, description string, answerOptions interface{}, endBlockHeight int64) int64 {
	validateMnemonic := validateMnemonic(mnemonic, starterAddress)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateStart := vote_con.ValidateStart(title, description, starterAddress, answerOptions, endBlockHeight)
	if validateStart != 0 {
		return validateStart
	}

	validateTxInMemory := validateTxInMemory(starterAddress, config.VoteSuperAddress, "vote_contract_start_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}
	return 0
}
