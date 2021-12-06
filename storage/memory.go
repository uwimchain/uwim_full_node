package storage

import (
	"log"
	"node/config"
	"node/memory"
	"node/storage/deep_actions"
)

var TransactionsMemory deep_actions.Txs

var BlockMemory Block

type Block struct {
	Height            int64              `json:"height"`
	PrevHash          string             `json:"prevHash"`
	Timestamp         string             `json:"timestamp"`
	Proposer          string             `json:"proposer"`
	ProposerSignature []byte             `json:"proposerSignature"`
	Body              deep_actions.Txs   `json:"body"`
	Votes             deep_actions.Votes `json:"votes"`
}

func AppendTxToTransactionMemory(tx deep_actions.Tx) {
	if memory.IsValidator() {
		TransactionsMemory = append(TransactionsMemory, tx)
	}
}

func ClearTransactionMemory() {
	if TransactionsMemory != nil {
		var clearedMemory deep_actions.Txs
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
