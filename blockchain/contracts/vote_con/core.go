package vote_con

import "node/blockchain/contracts"

var (
	db = contracts.Database{}

	VoteDB   = db.NewConnection("blockchain/contracts/vote_con/storage/vote_contract_vote")
	EventDB  = db.NewConnection("blockchain/contracts/vote_con/storage/vote_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/vote_con/storage/vote_contract_config")
)

type Vote struct {
	Nonce int64 `json:"nonce"`
	StartTimestamp contracts.String `json:"start_timestamp"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	AnswerOptions  []AnswerOption   `json:"answer_options"`
	Answers        []AddressAnswer  `json:"answers"`
	EndBlockHeight int64            `json:"end_block_height"`
	EndTimestamp contracts.String `json:"end_timestamp"`
	HardTimestamp contracts.String `json:"hard_timestamp"`
}

func NewVote(nonce int64, startTimestamp string, title string, description string, answerOptions []AnswerOption, answers []AddressAnswer, endBlockHeight int64, endTimestamp string, hardTimestamp string) *Vote {
	return &Vote{Nonce: nonce, StartTimestamp: contracts.String(startTimestamp), Title: title, Description: description,
		AnswerOptions: answerOptions, Answers: answers, EndBlockHeight: endBlockHeight, EndTimestamp: contracts.String(endTimestamp),
		HardTimestamp: contracts.String(hardTimestamp)}
}

type AnswerOption struct {
	PossibleAnswer      string  `json:"possible_answer"`
	PossibleAnswerNonce string  `json:"possible_answer_nonce"`
	AnswerCost          float64 `json:"answer_cost"`
}

type AddressAnswer struct {
	Address             string `json:"address"`
	Signature           []byte `json:"signature"`
	TxHash              string `json:"tx_hash"`
	BlockHeight         int64  `json:"block_height"`
	PossibleAnswerNonce string `json:"possible_answer_nonce"`
	VoteNonce int64 `json:"vote_nonce"`
}

func NewAddressAnswer(address string, signature []byte, txHash string, blockHeight int64, possibleAnswerNonce string, voteNonce int64) *AddressAnswer {
	return &AddressAnswer{Address: address, Signature: signature, TxHash: txHash, BlockHeight: blockHeight, PossibleAnswerNonce: possibleAnswerNonce, VoteNonce: voteNonce}
}

var VoteMemory []Vote
