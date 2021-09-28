package storage

import (
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
	t    deep_actions.Token
	a    deep_actions.Address
	tx   deep_actions.Tx
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
	} else {
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
				t.Amount, _ = apparel.Round(t.Amount)
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

		check := c.GetChain(strconv.FormatInt(config.BlockHeight, 10))
		if check != "" {
			ConfigUpdate("block_height", strconv.FormatInt(config.BlockHeight+1, 10))
		} else {
			log.Println("Storage core addBlock error 2: block don`t written")
		}
	}
}

func ConfigUpdate(parameter string, value string) {
	conf.ConfigUpdate(parameter, value)
}

func ZeroBlock() {
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
}

func CheckBlock(height int64) bool {
	return leveldb.ChainDB.Has(strconv.FormatInt(height, 10))
}

func NewBlocksForStart(blocks []deep_actions.Chain) {
	if memory.DownloadBlocks {

		if err := validateDownloadBlocks(blocks); err != nil {
			log.Println("Core new blocks for start error:", err)
		} else {

			for _, block := range blocks {

				err := c.NewChain(block)
				if err != nil {
					log.Println("Core new blocks for start error: ", err)
				} else {
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

					check := c.GetChain(strconv.FormatInt(config.BlockHeight, 10))
					if check != "" {
						ConfigUpdate("block_height", strconv.FormatInt(config.BlockHeight+1, 10))
						config.BlockHeight++
					} else {
						log.Println("Storage core NewBlocksForStart error: block don`t written")
					}
				}
			}

		}

		memory.DownloadBlocks = false

	}
}

func NewTx(transactionType int64, nonce int64, hashTx string, height int64, from string, to string, amount float64, tokenLabel string, timestamp string, tax float64, signature []byte, proposer string, comment deep_actions.Comment) {
	switch transactionType {
	case 1:
		{
			switch comment.Title {
			case "smart_contract_abandonment":
				{
					err := a.ScAbandonment(from)
					if err != nil {
						log.Println(err)
					}
					break
				}
			}

			break
		}
	case 3:
		{
			switch comment.Title {
			case "create_token_transaction":
				{
					if comment.Data == nil {
						log.Println("Deep actions new tx error 1: comment data is null")
						return
					}
					err := json.Unmarshal(comment.Data, &t)
					if err != nil {
						log.Println("Deep actions new tx error 11:", err)
						return
					} else {
						t.NewToken(t.Type, t.Label, t.Name, t.Proposer, t.Signature, t.Emission, t.Timestamp)
					}

					break
				}
			case "rename_token_transaction":
				{
					err := json.Unmarshal(comment.Data, &t)
					if err != nil {
						log.Println("Deep actions new tx error 2:", err)
						return
					} else {
						t.RenameToken(t.Label, t.Name)
					}
					break
				}
			case "change_token_standard_transaction":
				{
					err := json.Unmarshal(comment.Data, &t)
					if err != nil {
						log.Println("Deep actions new tx error 3:", err)
						return
					} else {
						t.ChangeTokenStandard(t.Label, t.Standard, timestamp, hashTx)
					}
					break
				}
			case "fill_token_card_transaction":
				{
					tokenLabel := GetAddressToken(from).Label
					t.FillTokenCard(tokenLabel, comment.Data, timestamp, hashTx)
					break
				}
			case "fill_token_standard_card_transaction":
				{
					tokenLabel := GetAddressToken(from).Label
					t.FillTokenStandardCard(tokenLabel, comment.Data, timestamp, hashTx)
					break
				}
			}

			break
		}
	}

	if from != "" && to != "" {
		amount, _ = apparel.Round(amount)
		tax, _ = apparel.Round(tax)
		if amount != 0 {
			a.UpdateBalance(from, *deep_actions.NewBalance(tokenLabel, amount, timestamp), false)

			a.UpdateBalance(to, *deep_actions.NewBalance(tokenLabel, amount, timestamp), true)
		}

		if tax != 0 {
			a.UpdateBalance(from, *deep_actions.NewBalance(config.BaseToken, tax, timestamp), false)

			if !memory.DownloadBlocks {
				a.UpdateBalance(proposer, *deep_actions.NewBalance(config.BaseToken, tax, timestamp), true)
			}
		}

		transaction := deep_actions.NewTx(transactionType, nonce, hashTx, height, from, to, amount, tokenLabel, timestamp, tax, signature, comment)

		rowFrom := tx.GetTx(from)
		rowTo := tx.GetTx(to)

		jsonString := deep_actions.AppendTx(rowFrom, *transaction)
		leveldb.TxDB.Put(from, jsonString)

		jsonString = deep_actions.AppendTx(rowTo, *transaction)
		leveldb.TxDB.Put(to, jsonString)

		if hashTx == "" {
			log.Println("HASH TX IS NULL!!!!")
		}
		leveldb.TxsDB.Put(hashTx, strconv.FormatInt(height, 10))
	} else {
		return
	}
}

func validateDownloadBlocks(blocks []deep_actions.Chain) error {
	return nil
}

func validateDownloadBlockTxs(txs []deep_actions.Tx) bool {
	for _, t := range txs {
		if leveldb.TxsDB.Has(t.HashTx) {
			return false
		}
	}

	return true
}

func GetBlockHash(height int64) string {
	Chain := c.GetChain(strconv.FormatInt(height, 10))
	Hash := ""
	if Chain != "" {
		err := json.Unmarshal([]byte(Chain), &c)
		if err != nil {
			log.Println("Get block hash error:", err)
		}
		Hash = c.Hash
	}

	return Hash
}

func GetBLockForHeight(height int64) string {
	return c.GetChain(strconv.FormatInt(height, 10))

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
	prevChain := c.GetChain(strconv.FormatInt(prevChainKey-1, 10))
	prevHash := ""
	if prevChain != "" {
		err := json.Unmarshal([]byte(prevChain), &c)
		if err != nil {
			log.Println("Get prev block hash error:", err)
		}

		prevHash = c.Hash
	}

	return prevHash
}

func GetBalance(address string) []deep_actions.Balance {
	row := a.GetAddress(address)
	Addr := deep_actions.Address{}

	if row != "" {
		err := json.Unmarshal([]byte(row), &Addr)
		if err != nil {
			log.Println("Get balance error:", err)
		}
	}

	return Addr.Balance
}

func GetBalanceForToken(address string, tokenLabel string) deep_actions.Balance {
	row := a.GetAddress(address)
	Addr := deep_actions.Address{}
	tokenBalance := deep_actions.Balance{}

	if row != "" {
		err := json.Unmarshal([]byte(row), &Addr)
		if err != nil {
			log.Println("Get balance error:", err)
		}
	}

	for _, i := range Addr.Balance {
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
		publicKey, _ := crypt.PublicKeyFromAddress(address)
		scAddress = crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)
	}

	if scAddress == "" {
		return nil
	}

	balance := GetBalance(scAddress)
	return balance
}

func GetAddressToken(address string) deep_actions.Token {
	row := a.GetAddress(address)
	token := deep_actions.Token{}

	if row != "" {
		Addr := deep_actions.Address{}
		err := json.Unmarshal([]byte(row), &Addr)
		if err != nil {
			log.Println("Get token from address error 1:", err)
		} else {

			row = GetTokenJson(Addr.TokenLabel)
			if row != "" {
				err := json.Unmarshal([]byte(row), &token)
				if err != nil {
					log.Println("Get token from address error 2:", err)
				}
			}
		}
	}

	return token
}

func GetAddress(address string) deep_actions.Address {
	row := a.GetAddress(address)
	addressData := deep_actions.Address{}
	if row != "" {
		err := json.Unmarshal([]byte(row), &addressData)
		if err != nil {
			fmt.Println("Get address data from address error 1:", err)
		}
	}

	return addressData
}

func CheckAddressToken(address string) bool {
	addressData := GetAddress(address)
	return addressData.TokenLabel != ""
}

func IsAddressTokenOwner(address string, tokenLabel string) bool {
	addressData := GetAddress(address)
	if addressData.TokenLabel == tokenLabel {
		return true
	}

	return false
}

func CheckAddressScKeeping(address string) bool {
	if crypt.IsAddressUw(address) {
		row := a.GetAddress(address)
		if row != "" {
			Addr := deep_actions.Address{}
			err := json.Unmarshal([]byte(row), &Addr)
			if err != nil {
				log.Println("Deep actions check address token error 1:", err)
				return true
			}
			return Addr.ScKeeping
		}
	}

	return true
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

func CheckAddress(address string) bool {
	return leveldb.AddressDB.Has(address)
}

func CalculateReward(address string) float64 {
	addressBalance := GetBalanceForToken(address, config.RewardTokenLabel)
	if config.BlockHeight > config.AnnualBlockHeight {
		return config.RewardCoefficientStage2
	} else {
		reward, _ := apparel.Round((addressBalance.Amount * config.RewardCoefficientStage1) / 100)
		return reward
	}
}

func GetTransactions(address string) []deep_actions.Tx {
	var result []deep_actions.Tx
	transactions := tx.GetTx(address)
	if transactions == "" {
		return nil
	} else {
		err := json.Unmarshal([]byte(transactions), &result)
		if err != nil {
			log.Println("get transactions error:", err)
			return nil
		} else {
			if int64(len(result)-1) <= config.BalanceTransactionsLimit {
				return reverseTxs(result)
			} else {
				return reverseTxs(result[int64(len(result))-config.BalanceTransactionsLimit:])
			}
		}
	}
}

func reverseTxs(txs []deep_actions.Tx) []deep_actions.Tx {
	if len(txs) == 0 {
		return txs
	}
	return append(reverseTxs(txs[1:]), txs[0])
}

func GetTxForHash(hash string) string {
	result := ""

	row := leveldb.TxsDB.Get(hash)
	if row.Value != "" {
		block := deep_actions.Chain{}

		height := apparel.ParseInt64(row.Value)
		jsonBlock := GetBLockForHeight(height)
		err := json.Unmarshal([]byte(jsonBlock), &block)
		if err != nil {
			log.Println("Get tx for hash error:", err)
		}

		if block.Txs != nil {
			for _, t := range block.Txs {
				if t.HashTx == hash {
					jsonString, err := json.Marshal(t)
					if err != nil {
						log.Println("Get tx for hash error:", err)
					}

					result = string(jsonString)
					break
				}
			}
		}
	}

	return result
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
		ZeroBlock()
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

func CheckToken(label string) bool {
	return t.CheckToken(label)
}

func GetTokens(start, limit int64) (interface{}, error) {
	var tokens []deep_actions.Token
	if start < 0 {
		tokens = t.GetAllTokens()
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

func GetTokenJson(label string) string {
	return t.GetToken(label)
}

func GetToken(label string) deep_actions.Token {
	jsonString := t.GetToken(label)

	token := deep_actions.Token{}
	_ = json.Unmarshal([]byte(jsonString), &token)

	return token
}

func TokenAbandonment(address string, label string) error {
	if !CheckToken(label) {
		return errors.New("storage token abandonment error 1: token does not exist")
	}

	row := a.GetAddress(address)
	Addr := deep_actions.Address{}
	err := json.Unmarshal([]byte(row), &Addr)
	if err != nil {
		return errors.New(fmt.Sprintf("Storage token abandonment error 2: %v", err))
	}

	if Addr.TokenLabel != label {
		return errors.New("Storage token abandonment error 3: this address dont have this token")
	}

	Addr.TokenLabel = ""

	jsonString, err := json.Marshal(Addr)
	if err != nil {
		return errors.New(fmt.Sprintf("Storage token abandonment error 4: %v", err))
	}

	leveldb.AddressDB.Put(address, string(jsonString))

	return nil
}
