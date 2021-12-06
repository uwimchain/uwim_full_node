package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
	"node/storage/deep_actions"
	"node/storage/leveldb"
	"sort"
	"strconv"
)

func Init() {
	BlockHeightUpdate()
}

func AddBlock() {
	votes := getBlockVotes()
	chainHeader := deep_actions.Header{
		PrevHash:          GetPrevBlockHash(),
		TxCounter:         int64(len(BlockMemory.Body)),
		Timestamp:         BlockMemory.Timestamp,
		ProposerSignature: BlockMemory.ProposerSignature,
		Proposer:          BlockMemory.Proposer,
		Votes:             votes,
		VoteCounter:       int64(len(votes)),
	}

	chain := deep_actions.NewChain(chainHeader, BlockMemory.Body)
	chain.SetHash()

	if err := chain.Update(); err != nil {
		log.Println(fmt.Sprintf("Storage core addBlock error 1: %v", err))
		return
	}

	proposersPubKey, _ := crypt.PublicKeyFromAddress(chain.Header.Proposer)
	proposerAddress := crypt.AddressFromPublicKey(metrics.AddressPrefix, proposersPubKey)

	for _, t := range chain.Txs {
		if CheckTx(t.HashTx) {
			log.Println("Storage core add block new tx error: this tx already exists")

			if TransactionsMemory != nil {
				for idx, t := range TransactionsMemory {
					if t.Nonce == t.Nonce {
						TransactionsMemory = append(
							TransactionsMemory[:idx],
							TransactionsMemory[idx+1:]...,
						)
						break
					}
				}
			}
		} else {
			t.Amount = apparel.Round(t.Amount)
			NewTx(
				t.Type,
				t.Nonce,
				t.HashTx,
				config.BlockHeight,
				t.From,
				t.To,
				t.Amount,
				t.TokenLabel,
				t.Timestamp,
				t.Tax,
				t.Signature,
				proposerAddress,
				t.Comment,
			)

			if t.Type != 2 {
				if TransactionsMemory != nil {
					for idx, t := range TransactionsMemory {
						if t.Nonce == t.Nonce {
							TransactionsMemory = append(
								TransactionsMemory[:idx],
								TransactionsMemory[idx+1:]...,
							)
							break
						}
					}
				}
			}
		}
	}

	deep_actions.ConfigUpdate("block_height", strconv.FormatInt(config.BlockHeight+1, 10))
}

func CheckBlock(height int64) bool {
	return leveldb.ChainDB.Has(strconv.FormatInt(height, 10))
}

func NewBlocksForStart(chains deep_actions.Chains) {
	if !memory.DownloadBlocks {
		return
	}

	for _, chain := range chains {
		_ = chain.Update()

		for _, t := range chain.Txs {
			NewTx(
				t.Type,
				t.Nonce,
				t.HashTx,
				t.Height,
				t.From,
				t.To,
				t.Amount,
				t.TokenLabel,
				t.Timestamp,
				t.Tax,
				t.Signature,
				chain.Header.Proposer,
				t.Comment,
			)

			if t.Type != 2 {
				if TransactionsMemory != nil {
					for idx, mTransaction := range TransactionsMemory {
						if mTransaction.Nonce == t.Nonce {
							TransactionsMemory = append(TransactionsMemory[:idx], TransactionsMemory[idx+1:]...)
							break
						}
					}
				}
			}
		}

		deep_actions.ConfigUpdate("block_height", strconv.FormatInt(config.BlockHeight+1, 10))
		config.BlockHeight++
	}

	memory.DownloadBlocks = false
}

func NewTx(transactionType int64, nonce int64, hashTx string, height int64, from string, to string, amount float64, tokenLabel string, timestamp string, tax float64, signature []byte, proposer string, comment deep_actions.Comment) {
	switch transactionType {
	case 1:
		switch comment.Title {
		case "smart_contract_abandonment":
			address := deep_actions.GetAddress(from)
			address.ScAbandonment()
			break
		}
		break
	case 3:
		commentData := make(map[string]interface{})
		_ = json.Unmarshal(comment.Data, &commentData)

		switch comment.Title {
		case "create_token_transaction":
			tokenStandard := apparel.ConvertInterfaceToInt64(commentData["standard"])
			tokenType := apparel.ConvertInterfaceToInt64(commentData["type"])

			signature, _ := base64.StdEncoding.DecodeString(apparel.ConvertInterfaceToString(commentData["signature"]))

			token := deep_actions.NewToken(tokenType,
				apparel.ConvertInterfaceToString(commentData["label"]),
				apparel.ConvertInterfaceToString(commentData["name"]),
				from, signature,
				apparel.ConvertInterfaceToFloat64(commentData["emission"]),
				timestamp,
				tokenStandard)

			token.Create()
			break
		case "rename_token_transaction":
			address := deep_actions.GetAddress(from)
			token := address.GetToken()
			token.RenameToken(apparel.ConvertInterfaceToString(commentData["new_name"]))
			break
		case "change_token_standard_transaction":
			address := deep_actions.GetAddress(from)
			token := address.GetToken()
			token.ChangeTokenStandard(apparel.ConvertInterfaceToInt64(commentData["standard"]), timestamp, hashTx)
			break
		case "fill_token_card_transaction":
			address := deep_actions.GetAddress(from)
			token := address.GetToken()
			token.FillTokenCard(comment.Data, timestamp, hashTx)
			break
		case "fill_token_standard_card_transaction":
			address := deep_actions.GetAddress(from)
			token := address.GetToken()
			token.FillTokenStandardCard(comment.Data, timestamp, hashTx)
			break
		}

		break
	}

	if from == "" || to == "" {
		return
	}
	amount = apparel.Round(amount)
	tax = apparel.Round(tax)

	addressFrom := deep_actions.GetAddress(from)
	addressTo := deep_actions.GetAddress(to)

	addressProposer := deep_actions.GetAddress(proposer)

	if amount != 0 {
		addressFrom.UpdateBalance(from, amount, tokenLabel, timestamp, false)

		addressTo.UpdateBalance(to, amount, tokenLabel, timestamp, true)
	}

	if tax != 0 {
		addressFrom.UpdateBalance(from, tax, config.BaseToken, timestamp, false)

		if !memory.DownloadBlocks {
			addressProposer.UpdateBalance(proposer, tax, config.BaseToken, timestamp, true)
		}
	}

	tx := deep_actions.Tx{
		Type:       transactionType,
		Nonce:      nonce,
		HashTx:     hashTx,
		Height:     height,
		From:       from,
		To:         to,
		Amount:     amount,
		TokenLabel: tokenLabel,
		Timestamp:  timestamp,
		Tax:        tax,
		Signature:  signature,
		Comment:    comment,
	}

	addressFrom.AppendTx(tx)
	addressTo.AppendTx(tx)

	leveldb.TxsDB.Put(hashTx, strconv.FormatInt(height, 10))
}

func getBlockVotes() deep_actions.Votes {
	var result deep_actions.Votes
	for _, vote := range BlockMemory.Votes {
		result = append(result, deep_actions.Vote{Proposer: vote.Proposer, Signature: vote.Signature,
			BlockHeight: vote.BlockHeight, Vote: vote.Vote})
	}

	return result
}

func GetPrevBlockHash() string {
	prevChainKey, _ := strconv.ParseInt(deep_actions.GetConfig("block_height"), 10, 64)
	prevChain := deep_actions.GetChain(prevChainKey - 1)

	return prevChain.Hash
}

func GetBalance(addressString string) []deep_actions.Balance {
	address := deep_actions.GetAddress(addressString)
	return address.Balance
}

func GetBalanceForToken(addressString string, tokenLabel string) deep_actions.Balance {
	address := deep_actions.GetAddress(addressString)
	tokenBalance := deep_actions.Balance{}

	for _, i := range address.Balance {
		if i.TokenLabel == tokenLabel {
			return i
		}
	}

	return tokenBalance
}

func GetScBalance(address string) []deep_actions.Balance {
	if address == "" {
		return nil
	}

	scAddress := address
	if !crypt.IsAddressSmartContract(address) {
		scAddress = crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, address)
	}

	if scAddress == "" {
		return nil
	}

	balance := GetBalance(scAddress)
	return balance
}

func GetAllNodesBalances() float64 {
	var result float64

	if memory.ValidatorsMemory == nil {
		return result
	}

	for _, i := range memory.ValidatorsMemory {
		balance := GetBalanceForToken(crypt.AddressFromAnotherAddress(metrics.NodePrefix, i.Address), config.RewardTokenLabel)
		result += balance.Amount
	}

	return result
}

func CalculateReward(address string) float64 {
	balance := GetBalance(address)
	if balance == nil {
		return 0
	}

	var amount float64 = 0
	for _, i := range balance {
		if i.TokenLabel == config.RewardTokenLabel {
			amount = i.Amount
			break
		}
	}

	if amount == 0 {
		return 0
	}

	var reward float64 = 0
	if config.BlockHeight > config.AnnualBlockHeight {
		emitRateIdx := config.GetEmitRateIdx()
		if emitRateIdx < 0 {
			return 0
		}

		reward = apparel.Round((amount*config.EmitRate[emitRateIdx])/1000000) * float64(len(memory.ValidatorsMemory))
	} else {
		reward = apparel.Round((amount * config.RewardCoefficientStage1) / 100)
	}

	return reward
}

func GetTxForHash(hash string) string {
	row := leveldb.TxsDB.Get(hash)
	if row.Value != "" {
		height, _ := strconv.ParseInt(row.Value, 10, 64)
		block := deep_actions.GetChain(height)

		if block.Txs != nil {
			for _, i := range block.Txs {
				if i.HashTx == hash {
					jsonString, err := json.Marshal(i)
					if err != nil {
						log.Println("Get tx for hash error:", err)
					}

					return string(jsonString)
				}
			}
		}
	}

	return ""
}

func CheckTx(hashTx string) bool {
	return leveldb.TxsDB.Has(hashTx)
}

func BlockHeightUpdate() {
	config.BlockHeight = GetBlockHeight()
}

func GetBlockHeight() int64 {
	result, _ := strconv.ParseInt(deep_actions.GetConfig("block_height"), 10, 64)

	return result
}

func GetTokenId() int64 {
	result, err := strconv.ParseInt(deep_actions.GetConfig("token_id"), 10, 64)
	if err != nil {
		log.Println("Get token id error:", err)
	}
	return result
}

func CheckToken(label string) bool {
	return deep_actions.CheckToken(label)
}

func GetTokens(start, limit int64) (interface{}, error) {
	var tokens []deep_actions.Token
	if start < 0 {
		tokens = deep_actions.GetAllTokens()
	} else {
		for i := start; i <= start+limit-1; i++ {
			token := GetTokenForId(i)

			if token.Id != 0 {
				tokens = append(tokens, *token)
			}
		}
	}

	if tokens == nil {
		return nil, errors.New("error 2: empty tokens list")
	} else {
		sort.Slice(tokens, func(i, j int) bool {
			return tokens[i].Id < tokens[j].Id
		})
	}

	return tokens, nil
}

func GetTokenForId(tokenId int64) *deep_actions.Token {
	tokenLabel := leveldb.TokenIdsDb.Get(strconv.FormatInt(tokenId, 10)).Value
	if tokenLabel != "" {
		return deep_actions.GetToken(tokenLabel)
	}

	return nil
}

func AddTokenEmission(tokenLabel string, addEmissionAmount float64) error {
	token := deep_actions.GetToken(tokenLabel)

	if token.Emission+addEmissionAmount > config.MaxEmission {
		return errors.New("add emission amount tran greater than max emission config parameter")
	}

	token.AddTokenEmission(addEmissionAmount)

	return nil
}
