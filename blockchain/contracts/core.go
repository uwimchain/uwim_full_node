package contracts

import (
	"node/config"
	"node/storage"
	"node/storage/deep_actions"
	"node/websocket/sender"
)

var (
	TransactionsMemory   = &storage.TransactionsMemory
	NewTx                = deep_actions.NewTx
	NewComment           = deep_actions.NewComment
	TokenAbandonment     = storage.TokenAbandonment
	StorageA             = deep_actions.Address{}
	StorageUpdateBalance = StorageA.UpdateBalance
	GetDelegateScBalance = storage.GetBalance(config.DelegateScAddress)
	NewBalance           = deep_actions.NewBalance
	SendTx               = sender.SendTx
)

type ContractCommentData struct {
	NodeAddress string `json:"node_address"`
	CheckSum    []byte `json:"check_sum"`
}

func NewContractCommentData(nodeAddress string, checkSum []byte) *ContractCommentData {
	return &ContractCommentData{
		NodeAddress: nodeAddress,
		CheckSum:    checkSum,
	}
}
