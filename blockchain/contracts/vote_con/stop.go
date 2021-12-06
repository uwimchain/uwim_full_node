package vote_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/blockchain/contracts"
	"node/config"
	"strconv"
)

type StopArgs struct {
	StopBlockHeight int64 `json:"stop_block_height"`
	StopTimestamp contracts.String `json:"stop_timestamp"`
}

func NewStopArgs(stopBlockHeight int64, stopTimestamp string) *StopArgs {
	return &StopArgs{StopBlockHeight: stopBlockHeight, StopTimestamp: contracts.String(stopTimestamp)}
}

func (args *StopArgs) Stop() error {
	err := stop(args.StopBlockHeight, string(args.StopTimestamp))
	if err != nil {
		return errors.New(fmt.Sprintf("stop error 1: %v", err))
	}

	return nil
}

func stop(blockHeight int64, timestamp string) error {
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

		voteJson := VoteDB.Get(strconv.FormatInt(VoteMemory[idx].Nonce, 10)).Value
		if voteJson == "" {
			return errors.New("error 1: vote does not exist in database")
		}

		err := json.Unmarshal([]byte(voteJson), &vote)
		if err != nil {
			return errors.New(fmt.Sprintf("erorr 2: %v", err))
		}

		if vote.Nonce == 0 {
			return errors.New("error 3: vote does not exist in database")
		}

		vote.EndTimestamp = contracts.String(timestamp)

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

		VoteDB.Put(strconv.FormatInt(VoteMemory[idx].Nonce, 10), string(jsonVote))
	}

	VoteMemory = newVoteMemory

	return nil
}
