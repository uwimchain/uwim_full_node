package deep_actions

import (
	"encoding/json"
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

func (tx *Tx) GetJsonForValidateSignature() []byte {
	jsonString, _ := json.Marshal(Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})

	return jsonString
}

type Chain struct {
	Hash   string `json:"hash"`
	Header Header `json:"header"`
	Txs    Txs    `json:"txs"`
}

type Chains []Chain

func NewChain(header Header, txs Txs) *Chain {
	return &Chain{
		Header: header,
		Txs:    txs,
	}
}

func (chain *Chain) SetHash() {
	jsonString, _ := json.Marshal(Chain{
		Header: Header{
			PrevHash:          chain.Header.PrevHash,
			TxCounter:         chain.Header.TxCounter,
			Timestamp:         chain.Header.Timestamp,
			ProposerSignature: chain.Header.ProposerSignature,
			Proposer:          chain.Header.Proposer,
			Votes:             nil,
			VoteCounter:       0,
		},
		Txs: chain.Txs,
	})

	chain.Hash = crypt.GetHash(jsonString)
}

type Header struct {
	PrevHash          string `json:"prevHash"`
	TxCounter         int64  `json:"txCounter"`
	Timestamp         string `json:"timestamp"`
	ProposerSignature []byte `json:"proposerSignature"`
	Proposer          string `json:"proposer"`
	Votes             Votes  `json:"votes"`
	VoteCounter       int64  `json:"voteCounter"`
}

type Vote struct {
	Proposer    string `json:"proposer"`
	Signature   []byte `json:"signature"`
	BlockHeight int64  `json:"blockHeight"`
	Vote        bool   `json:"vote"`
}

type Votes []Vote

func (chain *Chain) Update() error {
	jsonString, _ := json.Marshal(chain)

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

func GetConfig(key string) string {
	return leveldb.ConfigDB.Get(key).Value
}

func ConfigUpdate(key string, value string) {
	leveldb.ConfigDB.Put(key, value)
}
