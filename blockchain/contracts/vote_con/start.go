package vote_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"strconv"
)

type StartArgs struct {
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	AnswerOptions  []AnswerOption `json:"answer_options"`
	EndBlockHeight int64          `json:"end_block_height"`
	StarterAddress string         `json:"starter_address"`
	TxHash         string         `json:"tx_hash"`
	BlockHeight    int64          `json:"block_height"`
}

func NewStartArgs(title string, description string, answerOptions interface{}, endBlockHeight int64, starterAddress string, txHash string, blockHeight int64) (*StartArgs, error) {
	var answerOptionsData []AnswerOption
	answerOptionsJson, err := json.Marshal(answerOptions)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("new start args error 1: %v", err))
	}

	err = json.Unmarshal(answerOptionsJson, &answerOptionsData)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("new start args error 2: %v", err))
	}

	return &StartArgs{Title: title, Description: description, AnswerOptions: answerOptionsData,
		EndBlockHeight: endBlockHeight, StarterAddress: starterAddress, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func Start(args *StartArgs) error {
	err := start(args.Title, args.Description, args.TxHash, args.StarterAddress, args.AnswerOptions, args.EndBlockHeight, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("start vote error 1: %v", err))
	}
	return nil
}

func start(title, description, txHash, starterAddress string, answerOptions []AnswerOption, endBlockHeight, blockHeight int64) error {
	if starterAddress != config.VoteSuperAddress {
		return errors.New("error 1: permission denied")
	}

	if title == "" {
		return errors.New("error 2: title is empty")
	}

	if description == "" {
		return errors.New("error 3: description is empty")
	}

	if answerOptions == nil {
		return errors.New("error 4: answer options is empty")
	}

	if len(answerOptions) > config.MaxVoteAnswerOptions {
		return errors.New(fmt.Sprintf("error 5: count of answer options is more than %v", config.MaxVoteAnswerOptions))
	}

	if endBlockHeight <= config.BlockHeight {
		return errors.New("error 6: incorrect end block height")
	}

	if VoteMemory != nil {
		if len(VoteMemory) >= config.MaxVoteMemory {
			return errors.New("error 7: vote memory is full")
		}
	}

	for idx := range answerOptions {
		timestamp := apparel.TimestampUnix()

		if answerOptions[idx].AnswerCost < config.VoteAnswerOptionDefaultCost {
			answerOptions[idx].AnswerCost = config.VoteAnswerOptionDefaultCost
		}

		answerOptions[idx].PossibleAnswerNonce = strconv.FormatInt(apparel.GetNonce(strconv.FormatInt(timestamp, 10)), 10)
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
	vote := Vote{
		Nonce:          apparel.GetNonce(timestamp),
		StartTimestamp: contracts.String(timestamp),
		Title:          title,
		Description:    description,
		AnswerOptions:  answerOptions,
		Answers:        nil,
		EndBlockHeight: endBlockHeight,
		EndTimestamp:   "",
		HardTimestamp:  "",
	}

	jsonVote, err := json.Marshal(vote)
	if err != nil {
		return errors.New(fmt.Sprintf("error 8: %v", err))
	}

	err = contracts.AddEvent(config.VoteScAddress, *contracts.NewEvent("start", timestamp, blockHeight, txHash, starterAddress, newEventStartTypeData(title, description, starterAddress, answerOptions, endBlockHeight)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 9: %v", err))
	}

	VoteMemory = append(VoteMemory, vote)
	VoteDB.Put(strconv.FormatInt(vote.Nonce, 10), string(jsonVote))
	return nil
}
