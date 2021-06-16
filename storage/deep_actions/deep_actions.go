package deep_actions

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/config"
	"node/crypt"
	"node/storage/leveldb"
	"strconv"
)

var (
	A Address
)

type Address struct {
	Address     string    `json:"address"`
	Balance     []Balance `json:"balance"`
	PublicKey   []byte    `json:"publicKey"`
	FirstTxTime string    `json:"firstTxTime"`
	LastTxTime  string    `json:"lastTxTime"`
	TokenLabel  string    `json:"tokenLabel"`
	ScKeeping   bool      `json:"sc_keeping"`
}

func NewAddress(address string, balance []Balance, publicKey []byte, firstTxTime string, lastTxTime string,
	tokenLabel string) *Address {
	return &Address{
		Address:     address,
		Balance:     balance,
		PublicKey:   publicKey,
		FirstTxTime: firstTxTime,
		LastTxTime:  lastTxTime,
		TokenLabel:  tokenLabel,
		ScKeeping:   false,
	}
}

type Balance struct {
	TokenLabel string  `json:"tokenLabel"`
	Amount     float64 `json:"amount"`
	UpdateTime string  `json:"updateTime"`
}

func NewBalance(tokenLabel string, amount float64, updateTime string) *Balance {
	return &Balance{TokenLabel: tokenLabel, Amount: amount, UpdateTime: updateTime}
}

func (a *Address) NewAddress(address string, balance []Balance, publicKey []byte, firstTxTime string,
	lastTxTime string) {
	jsonString, err := json.Marshal(NewAddress(address, balance, publicKey, firstTxTime, lastTxTime, ""))
	if err != nil {
		log.Println("New Address error: ", err)
	}

	leveldb.AddressDB.Put(address, string(jsonString))
}

func (a *Address) GetAddress(address string) string {
	return leveldb.AddressDB.Get(address).Value
}

func (a *Address) UpdateBalance(address string, balance Balance, side bool) {
	row := a.GetAddress(address)

	if row == "" {
		publicKey, err := crypt.PublicKeyFromAddress(address)
		if err != nil {
			log.Println("Update Balance error 1:", err)
		}

		a.NewAddress(address, nil, publicKey, balance.UpdateTime, balance.UpdateTime)
	}

	row = a.GetAddress(address)
	Addr := Address{}
	err := json.Unmarshal([]byte(row), &Addr)
	if err != nil {
		log.Println("Update Balance error 2:", err)
	}

	Addr.Balance = updateBalance(Addr.Balance, balance, side)

	Addr.LastTxTime = balance.UpdateTime

	jsonString, err := json.Marshal(Addr)
	if err != nil {
		log.Println("Update Balance error 3:", err)
	}

	leveldb.AddressDB.Put(address, string(jsonString))
}

func (a *Address) CheckAddressToken(address string) bool {
	row := a.GetAddress(address)
	if row != "" {
		Addr := Address{}
		err := json.Unmarshal([]byte(row), &Addr)
		if err != nil {
			log.Println("Deep actions check address token error 1:", err)
			return false
		}
		return Addr.TokenLabel != ""
	}

	return false
}

func (a *Address) ScAbandonment(address string) error {

	if !crypt.IsAddressUw(address) || address == config.GenesisAddress {
		return errors.New("Deep actions smart contract abandonment error 1")
	}

	row := a.GetAddress(address)
	if row != "" {
		Addr := Address{}
		err := json.Unmarshal([]byte(row), &Addr)
		if err != nil {
			return errors.New("Deep actions smart contract abandonment error 2")
		}

		if Addr.ScKeeping {
			return errors.New("Deep actions smart contract abandonment error 3")
		}

		Addr.ScKeeping = true
		jsonString, err := json.Marshal(Addr)
		if err != nil {
			return errors.New("Deep actions smart contract abandonment error 4")
		}

		leveldb.AddressDB.Put(address, string(jsonString))
		return nil
	}

	return errors.New("Deep actions smart contract abandonment error 5")
}

type Tx struct {
	Type       int64   `json:"type"`
	Nonce      int64   `json:"nonce"`
	HashTx     string  `json:"hashTx"`
	Height     int64   `json:"height"`
	From       string  `json:"from"`
	To         string  `json:"to"`
	Amount     float64 `json:"amount"`
	TokenLabel string  `json:"tokenLabel"`
	Timestamp  string  `json:"timestamp"`
	Tax        float64 `json:"tax"`
	Signature  []byte  `json:"signature"`
	Comment    Comment `json:"comment"`
}

type Comment struct {
	Title string `json:"title"`
	Data  []byte `json:"data"`
}

func NewComment(title string, data []byte) *Comment {
	return &Comment{
		Title: title,
		Data:  data,
	}
}

func NewTx(txType int64, nonce int64, hashTx string, height int64, from string, to string, amount float64, tokenLabel string, timestamp string, tax float64, signature []byte, comment Comment) *Tx {
	return &Tx{
		Type:       txType,
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
}

func (tx *Tx) GetTx(address string) string {
	return leveldb.TxDB.Get(address).Value
}

type Chain struct {
	Hash   string `json:"hash"`
	Header Header `json:"header"`
	Txs    []Tx   `json:"txs"`
}

func NewChain(hash string, header Header, txs []Tx) *Chain {
	return &Chain{
		Hash:   hash,
		Header: header,
		Txs:    txs,
	}
}

type Header struct {
	PrevHash          string `json:"prevHash"`
	TxCounter         int64  `json:"txCounter"`
	Timestamp         string `json:"timestamp"`
	ProposerSignature []byte `json:"proposerSignature"`
	Proposer          string `json:"proposer"`
	Votes             []Vote `json:"votes"`
	VoteCounter       int64  `json:"voteCounter"`
}

func NewHeader(prevHash string, txCounter int64, timestamp string, proposerSignature []byte, proposer string, votes []Vote, voteCounter int64) *Header {
	return &Header{
		PrevHash:          prevHash,
		TxCounter:         txCounter,
		Timestamp:         timestamp,
		ProposerSignature: proposerSignature,
		Proposer:          proposer,
		Votes:             votes,
		VoteCounter:       voteCounter,
	}
}

type Vote struct {
	Proposer    string `json:"proposer"`
	Signature   []byte `json:"signature"`
	BlockHeight int64  `json:"blockHeight"`
	Vote        bool   `json:"vote"`
}

func NewVote(proposer string, signature []byte, blockHeight int64, vote bool) *Vote {
	return &Vote{
		Proposer:    proposer,
		Signature:   signature,
		BlockHeight: blockHeight,
		Vote:        vote,
	}
}

func (c *Chain) NewChain(chain Chain) error {
	chainForJson := Chain{
		Header: Header{
			PrevHash:          chain.Header.PrevHash,
			TxCounter:         chain.Header.TxCounter,
			Timestamp:         chain.Header.Timestamp,
			ProposerSignature: chain.Header.ProposerSignature,
			Proposer:          chain.Header.Proposer,
		},
		Txs: chain.Txs,
	}

	jsonForHash, err := json.Marshal(chainForJson)
	if err != nil {
		log.Println("New Chain error: ", err)
	}

	hash := crypt.GetHash(jsonForHash)

	if hash != chain.Header.PrevHash {

		jsonString, err := json.Marshal(NewChain(hash, chain.Header, chain.Txs))
		if err != nil {
			return errors.New(fmt.Sprintf("New Chain error: %v", err))
		}

		leveldb.ChainDB.Put(strconv.FormatInt(config.BlockHeight, 10), string(jsonString))
	} else {
		return errors.New("New chain error: hash == prev hash")
		//log.Println("New chain error: hash == prev hash")
	}

	return nil
}

func (c *Chain) GetChain(height string) string {
	return leveldb.ChainDB.Get(height).Value
}

type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (c *Config) GetConfig(key string) string {
	return leveldb.ConfigDB.Get(key).Value
}

func (c *Config) ConfigUpdate(key string, value string) {
	leveldb.ConfigDB.Put(key, value)
}

func findTokenInBalance(balance []Balance, token string) int {
	for i := range balance {
		if balance[i].TokenLabel == token {
			return i
		}
	}

	return len(balance)
}

func updateBalance(balance []Balance, newBalance Balance, side bool) []Balance {
	if idx := findTokenInBalance(balance, newBalance.TokenLabel); idx != len(balance) {
		if side {
			balance[idx].Amount += newBalance.Amount
			balance[idx].UpdateTime = newBalance.UpdateTime
		} else {
			if balance[idx].Amount < newBalance.Amount {
				return balance
			} else {
				balance[idx].Amount -= newBalance.Amount
				balance[idx].UpdateTime = newBalance.UpdateTime
			}
		}
	} else {
		balance = append(balance, newBalance)
	}

	return balance
}

func AppendTx(addressTxsRow string, tx Tx) string {
	var result []byte
	var AddressTxs []Tx
	if addressTxsRow == "" {
		AddressTxs = append(AddressTxs, tx)
	} else {
		err := json.Unmarshal([]byte(addressTxsRow), &AddressTxs)
		if err != nil {
			log.Println("append Tx error:", err)
		}

		AddressTxs = append(AddressTxs, tx)
	}

	result, err := json.Marshal(AddressTxs)
	if err != nil {
		log.Println("append Tx error:", err)
	}

	return string(result)
}
