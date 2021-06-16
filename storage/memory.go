package storage

import (
	"log"
	"node/config"
	"node/storage/deep_actions"
)

var TransactionsMemory []deep_actions.Tx

var BlockMemory Block

type Block struct {
	Height            int64               `json:"height"`
	PrevHash          string              `json:"prevHash"`
	Timestamp         string              `json:"timestamp"`
	Proposer          string              `json:"proposer"`
	ProposerSignature []byte              `json:"proposerSignature"`
	Body              []deep_actions.Tx   `json:"body"`
	Votes             []deep_actions.Vote `json:"votes"`
}

func NewBlock(height int64, prevHash string, timestamp string, proposer string, proposerSignature []byte,
	body []deep_actions.Tx, votes []deep_actions.Vote) *Block {
	return &Block{
		Height:            height,
		PrevHash:          prevHash,
		Timestamp:         timestamp,
		Proposer:          proposer,
		ProposerSignature: proposerSignature,
		Body:              body,
		Votes:             votes,
	}
}

func Update() {
	cleanMemory()
}

func cleanMemory() {
	if TransactionsMemory != nil {
		var clearedMemory []deep_actions.Tx
		for _, transaction := range TransactionsMemory {
			if (transaction.Height + config.StorageMemoryLifeIter) >= config.BlockHeight {
				clearedMemory = append(clearedMemory, transaction)
			}
		}

		if len(TransactionsMemory) != len(clearedMemory) {
			log.Println("Before cleaning memory:", len(TransactionsMemory))
			TransactionsMemory = clearedMemory
			log.Println("After cleaning memory:", len(TransactionsMemory))
		}
	}
}

func FindTxInMemory(nonce int64) bool {
	for _, transaction := range TransactionsMemory {
		if transaction.Nonce == nonce {
			return true
		}
	}
	return false
}
