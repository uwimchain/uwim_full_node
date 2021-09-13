package vote_con

import "node/blockchain/contracts"

var (
	db = contracts.Database{}

	VoteDB   = db.NewConnection("blockchain/contracts/vote_con/storage/vote_contract_vote")
	EventDB  = db.NewConnection("blockchain/contracts/vote_con/storage/vote_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/vote_con/storage/vote_contract_config")
)

type Vote struct {
	Nonce          string           `json:"nonce"`
	StartTimestamp int64           `json:"start_timestamp"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	AnswerOptions  []AnswerOption  `json:"answer_options"`
	Answers        []AddressAnswer `json:"answers"`
	EndBlockHeight int64           `json:"end_block_height"`
	EndTimestamp   int64           `json:"end_timestamp"`
	HardTimestamp  int64           `json:"hard_timestamp"`
}

func NewVote(nonce string, startTimestamp int64, title string, description string, answerOptions []AnswerOption, answers []AddressAnswer, endBlockHeight int64, endTimestamp int64, hardTimestamp int64) *Vote {
	return &Vote{Nonce: nonce, StartTimestamp: startTimestamp, Title: title, Description: description, AnswerOptions: answerOptions, Answers: answers, EndBlockHeight: endBlockHeight, EndTimestamp: endTimestamp, HardTimestamp: hardTimestamp}
}

type AnswerOption struct {
	PossibleAnswer      string  `json:"possible_answer"`
	PossibleAnswerNonce string   `json:"possible_answer_nonce"`
	AnswerCost          float64 `json:"answer_cost"`
}

type AddressAnswer struct {
	Address             string `json:"address"`
	Signature           []byte `json:"signature"`
	TxHash              string `json:"tx_hash"`
	BlockHeight         int64  `json:"block_height"`
	PossibleAnswerNonce string  `json:"possible_answer_nonce"`
	VoteNonce           string  `json:"vote_nonce"`
}

func NewAddressAnswer(address string, signature []byte, txHash string, blockHeight int64, possibleAnswerNonce string, voteNonce string) *AddressAnswer {
	return &AddressAnswer{Address: address, Signature: signature, TxHash: txHash, BlockHeight: blockHeight, PossibleAnswerNonce: possibleAnswerNonce, VoteNonce: voteNonce}
}

var VoteMemory []Vote
