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

var (
	c    deep_actions.Chain
	conf deep_actions.Config
)

func Init() {
	BlockHeightUpdate()
}

func AddBlock() {

	block := deep_actions.Chain{}

	for idx := range BlockMemory.Body {
		jsonForHash, _ := json.Marshal(BlockMemory.Body[idx])
		BlockMemory.Body[idx].HashTx = crypt.GetHash(jsonForHash)
		block.Txs = append(block.Txs, BlockMemory.Body[idx])
	}

	votes := getBlockVotes()
	block.Header = *deep_actions.NewHeader(
		GetPrevBlockHash(),
		int64(len(block.Txs)),
		BlockMemory.Timestamp,
		BlockMemory.ProposerSignature,
		BlockMemory.Proposer,
		votes,
		int64(len(votes)),
	)

	err := c.NewChain(block)
	if err != nil {
		log.Println(fmt.Sprintf("Storage core addBlock error 1: %v", err))
		return
	}
	proposersPubKey, _ := crypt.PublicKeyFromAddress(block.Header.Proposer)
	proposerAddress := crypt.AddressFromPublicKey(metrics.AddressPrefix, proposersPubKey)

	for _, t := range block.Txs {

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

//func ConfigUpdate(parameter string, value string) {
//	conf.ConfigUpdate(parameter, value)
//}

/*func ZeroBlock() {
	if memory.IsMainNode() && config.BlockHeight == 0 {
		if !CheckBlock(0) {
			var body []deep_actions.Tx

			timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)
			GenesisTransaction := deep_actions.NewTx(
				1,
				apparel.GetNonce(timestampD),
				"",
				0,
				"",
				config.GenesisAddress,
				1000000000,
				"uwm",
				timestampD,
				0,
				nil,
				*deep_actions.NewComment(
					"DefaultTransaction",
					nil,
				),
			)
			jsonForHash, err := json.Marshal(GenesisTransaction)
			if err != nil {
				log.Fatal("Zero block genesis transaction error:", err)
			}
			GenesisTransaction.HashTx = crypt.GetHash(jsonForHash)

			timestampD = strconv.FormatInt(apparel.TimestampUnix(), 10)
			MainTransaction := deep_actions.NewTx(
				1,
				apparel.GetNonce(timestampD),
				"",
				0,
				config.GenesisAddress,
				config.NodeNdAddress,
				326348839,
				"uwm",
				timestampD,
				0,
				crypt.SignMessageWithSecretKey(
					config.GenesisSecretKey,
					[]byte(config.GenesisAddress),
				),
				*deep_actions.NewComment(
					"DefaultTransaction",
					nil,
				),
			)
			jsonForHash, err = json.Marshal(MainTransaction)
			if err != nil {
				log.Fatal("Zero block main transaction error:", err)
			}
			MainTransaction.HashTx = crypt.GetHash(jsonForHash)

			timestampD = strconv.FormatInt(apparel.TimestampUnix(), 10)
			jsonString, err := json.Marshal(deep_actions.NewToken(
				1,
				0,
				"uwm",
				"UWM",
				config.GenesisAddress,
				crypt.SignMessageWithSecretKey(
					config.GenesisSecretKey,
					[]byte(config.GenesisAddress),
				),
				1000000000,
				apparel.TimestampUnix(),
				0,
			))
			if err != nil {
				log.Fatal("Zero block zero token transaction comment error:", err)
			}
			ZeroTokenTransaction := deep_actions.NewTx(
				3,
				apparel.GetNonce(timestampD),
				"",
				0,
				config.GenesisAddress,
				config.NodeNdAddress,
				1,
				"uwm",
				timestampD,
				0,
				crypt.SignMessageWithSecretKey(
					config.GenesisSecretKey,
					[]byte(config.GenesisAddress),
				),
				*deep_actions.NewComment(
					"CreateTokenTransaction",
					jsonString,
				),
			)
			jsonForHash, err = json.Marshal(ZeroTokenTransaction)
			if err != nil {
				log.Fatal("Zero block zero token transaction error:", err)
			}
			ZeroTokenTransaction.HashTx = crypt.GetHash(jsonString)

			body = append(body, *GenesisTransaction, *MainTransaction, *ZeroTokenTransaction)
			BlockMemory = *NewBlock(
				0,
				"",
				strconv.FormatInt(apparel.TimestampUnix(), 10),
				config.NodeNdAddress,
				crypt.SignMessageWithSecretKey(
					config.NodeSecretKey,
					[]byte(config.NodeNdAddress),
				),
				body,
				nil,
			)

			AddBlock()

			BlockMemory = Block{}

			log.Println("Zero block was witted")
		}
	}
}*/

func CheckBlock(height int64) bool {
	return leveldb.ChainDB.Has(strconv.FormatInt(height, 10))
}

func NewBlocksForStart(blocks []deep_actions.Chain) {
	if !memory.DownloadBlocks {
		return
	}

	if err := validateDownloadBlocks(blocks); err != nil {
		log.Println("Core new blocks for start error:", err)
		return
	}

	for _, block := range blocks {
		err := c.NewChain(block)
		if err != nil {
			log.Println("Core new blocks for start error: ", err)
			return
		}

		for _, t := range block.Txs {
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
				block.Header.Proposer,
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
		addressFrom.UpdateBalance(from, *deep_actions.NewBalance(tokenLabel, amount, timestamp), false)

		addressTo.UpdateBalance(to, *deep_actions.NewBalance(tokenLabel, amount, timestamp), true)
	}

	if tax != 0 {
		addressFrom.UpdateBalance(from, *deep_actions.NewBalance(config.BaseToken, tax, timestamp), false)

		if !memory.DownloadBlocks {
			addressProposer.UpdateBalance(proposer, *deep_actions.NewBalance(config.BaseToken, tax, timestamp), true)
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

func validateDownloadBlocks(blocks []deep_actions.Chain) error {

	return nil
}

func getBlockVotes() []deep_actions.Vote {
	var result []deep_actions.Vote
	for _, vote := range BlockMemory.Votes {
		result = append(result, *deep_actions.NewVote(vote.Proposer, vote.Signature, vote.BlockHeight, vote.Vote))
	}

	return result
}

func GetPrevBlockHash() string {
	prevChainKey, _ := strconv.ParseInt(conf.GetConfig("block_height"), 10, 64)
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

func CheckAddressToken(address string) bool {
	addressData := deep_actions.GetAddress(address)
	return addressData.TokenLabel != ""
}

func GetAllNodesBalances() float64 {
	rows := leveldb.AddressDB.GetAll(metrics.NodePrefix)
	var result float64
	for _, row := range rows {
		a := deep_actions.Address{}
		if err := json.Unmarshal([]byte(row.Value), &a); err == nil {
			for _, item := range a.Balance {
				if item.TokenLabel == config.RewardTokenLabel {
					result += item.Amount
					break
				}
			}
		} else {
			log.Println("Get all nodes balances error:", err)
		}
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
		height := apparel.ParseInt64(row.Value)
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

func GetConfig(key string) string {
	return conf.GetConfig(key)
}

func GetBlockHeight() int64 {
	result, _ := strconv.ParseInt(GetConfig("block_height"), 10, 64)

	if result == 0 {
		//ZeroBlock()
	}

	return result
}

func GetTokenId() int64 {
	result, err := strconv.ParseInt(GetConfig("token_id"), 10, 64)
	if err != nil {
		log.Println("Get token id error:", err)
	}
	return result
}

func GetTokens(start, limit int64) (interface{}, error) {
	var tokens []deep_actions.Token
	if start < 0 {
		tokens = deep_actions.GetAllTokens()
	} else {
		for i := start; i <= start+limit-1; i++ {
			token, err := GetTokenForId(i)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error 1: get token for id %v", err))
			}

			tokens = append(tokens, token)
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

func GetTokensCount() int64 {
	return leveldb.TokenDb.Count()
}

func GetTokenForId(tokenId int64) (deep_actions.Token, error) {
	var token deep_actions.Token
	tokenLabel := leveldb.TokenIdsDb.Get(strconv.FormatInt(tokenId, 10)).Value
	if tokenLabel != "" {
		tokenJson := leveldb.TokenDb.Get(tokenLabel).Value
		if tokenJson != "" {
			err := json.Unmarshal([]byte(tokenJson), &token)
			if err != nil {
				return token, errors.New(fmt.Sprintf("error 1: %v", err))
			}
		}
	}

	return token, nil
}

func AddTokenEmission(tokenLabel string, addEmissionAmount float64) error {
	token := deep_actions.GetToken(tokenLabel)

	if token.Emission+addEmissionAmount > config.MaxEmission {
		return errors.New("add emission amount tran greater than max emission config parameter")
	}

	token.AddTokenEmission(addEmissionAmount)

	return nil
}
