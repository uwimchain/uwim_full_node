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

// VoteContractAnswer method arguments
type VoteContractAnswerArgs struct {
	Mnemonic            string `json:"mnemonic"`
	PossibleAnswerNonce string `json:"possible_answer_nonce"`
	VoteNonce           string `json:"vote_nonce"`
}

func (api *Api) VoteContractAnswer(args *VoteContractAnswerArgs, result *string) error {
	args.Mnemonic = apparel.TrimToLower(args.Mnemonic)

	address := crypt.AddressFromMnemonic(args.Mnemonic)
	validateAnswer := validateAnswer(args.Mnemonic, address, args.VoteNonce, args.PossibleAnswerNonce)
	if validateAnswer != 0 {
		return errors.New(strconv.FormatInt(validateAnswer, 10))
	}

	commentData := make(map[string]interface{})
	commentData["possible_answer_nonce"] = args.PossibleAnswerNonce
	commentData["vote_nonce"] = args.VoteNonce

	commentDataJson, err := json.Marshal(commentData)
	if err != nil {
		log.Println("Send Transaction error 1:", err)
	}

	timestamp := apparel.TimestampUnix()

	var (
		txAmount float64 = 0
	)
	vote := vote_con.GetVoteForNonce(args.VoteNonce)
	if vote.Nonce != "" {
		if vote.AnswerOptions != nil {
			for _, i := range vote.AnswerOptions {
				if i.PossibleAnswerNonce == args.PossibleAnswerNonce {
					txAmount = i.AnswerCost
				}
			}
		}
	}

	comment := deep_actions.Comment{
		Title: "vote_contract_answer_transaction",
		Data:  commentDataJson,
	}

	secretKey := crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(args.Mnemonic))

	tx := deep_actions.Tx{
		Type:       1,
		Nonce:      apparel.GetNonce(strconv.FormatInt(timestamp, 10)),
		HashTx:     "",
		Height:     config.BlockHeight,
		From:       address,
		To:         config.VoteScAddress,
		Amount:     txAmount,
		TokenLabel: config.BaseToken,
		Timestamp:  strconv.FormatInt(timestamp, 10),
		Tax:        apparel.CalcTax(txAmount),
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

func validateAnswer(mnemonic, address string, voteNonce, possibleAnswerNonce string) int64 {
	validateMnemonic := validateMnemonic(mnemonic, address)
	if validateMnemonic != 0 {
		return validateMnemonic
	}

	validateAnswer := vote_con.ValidateAnswer(address, voteNonce, possibleAnswerNonce)
	if validateAnswer != 0 {
		return validateAnswer
	}

	validateTxInMemory := validateTxInMemory(address, config.VoteSuperAddress, "vote_contract_answer_transaction", 1)
	if validateTxInMemory != 0 {
		return validateTxInMemory
	}
	return 0
}
