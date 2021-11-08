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

type Txs []Tx

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

func GetTxJson(address string) string {
	return leveldb.TxDB.Get(address).Value
}

func (tx *Tx) SetSignature(secretKey []byte) {
	jsonString, _ := json.Marshal(Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})
	tx.Signature = crypt.SignMessageWithSecretKey(secretKey, jsonString)
}

func (tx *Tx) SetHash() {
	jsonString, _ := json.Marshal(tx)
	tx.HashTx = crypt.GetHash(jsonString)
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

	if hash == chain.Header.PrevHash {
		return errors.New("New chain error: hash == prev hash")
	}

	jsonString, err := json.Marshal(NewChain(hash, chain.Header, chain.Txs))
	if err != nil {
		return errors.New(fmt.Sprintf("New Chain error: %v", err))
	}

	leveldb.ChainDB.Put(strconv.FormatInt(config.BlockHeight, 10), string(jsonString))

	return nil
}

func GetChainJson(height int64) string {
	return leveldb.ChainDB.Get(strconv.FormatInt(height, 10)).Value
}

func GetChain(height int64) *Chain {
	chainJson := leveldb.ChainDB.Get(strconv.FormatInt(height, 10)).Value
	chain := Chain{}
	_ = json.Unmarshal([]byte(chainJson), &chain)

	return &chain
}

type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (c *Config) GetConfig(key string) string {
	return leveldb.ConfigDB.Get(key).Value
}

func ConfigUpdate(key string, value string) {
	leveldb.ConfigDB.Put(key, value)
}
