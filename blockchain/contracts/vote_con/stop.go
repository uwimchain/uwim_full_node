package vote_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/blockchain/contracts"
	"node/config"
)

type StopArgs struct {
	StopBlockHeight int64 `json:"stop_block_height"`
	StopTimestamp   int64 `json:"stop_timestamp"`
}

func NewStopArgs(stopBlockHeight int64, stopTimestamp int64) *StopArgs {
	return &StopArgs{StopBlockHeight: stopBlockHeight, StopTimestamp: stopTimestamp}
}

func Stop(args *StopArgs) error {
	err := stop(args.StopBlockHeight, args.StopTimestamp)
	if err != nil {
		return errors.New(fmt.Sprintf("stop error 1: %v", err))
	}

	return nil
}

func stop(blockHeight, timestamp int64) error {
	if VoteMemory == nil {
		return nil
	}

	var newVoteMemory []Vote

	for idx := range VoteMemory {
		if VoteMemory[idx].EndBlockHeight > blockHeight {
			newVoteMemory = append(newVoteMemory, VoteMemory[idx])
			continue
		}

		var vote Vote

		voteJson := VoteDB.Get(VoteMemory[idx].Nonce).Value
		if voteJson == "" {
			return errors.New("error 1: vote does not exist in database")
		}

		err := json.Unmarshal([]byte(voteJson), &vote)
		if err != nil {
			return errors.New(fmt.Sprintf("erorr 2: %v", err))
		}

		if vote.Nonce == "" {
			return errors.New("error 3: vote does not exist in database")
		}

		vote.EndTimestamp = timestamp

		jsonVote, err := json.Marshal(vote)
		if err != nil {
			return errors.New(fmt.Sprintf("error 4: %v", err))
		}

		err = contracts.AddEvent(config.VoteScAddress,
			*contracts.NewEvent("Stop", timestamp, blockHeight, "", "", nil),
			EventDB, ConfigDB)
		if err != nil {
			return errors.New(fmt.Sprintf("error 5: %v", err))
		}

		VoteDB.Put(VoteMemory[idx].Nonce, string(jsonVote))
	}

	VoteMemory = newVoteMemory

	return nil
}
