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

type HardStopArgs struct {
	VoteNonce      int64  `json:"vote_nonce"`
	TxHash         string `json:"tx_hash"`
	BlockHeight    int64  `json:"block_height"`
	StopperAddress string `json:"stopper_address"`
}

func NewHardStopArgs(voteNonce int64, txHash string, blockHeight int64, stopperAddress string) (*HardStopArgs, error) {
	return &HardStopArgs{VoteNonce: voteNonce, TxHash: txHash, BlockHeight: blockHeight,
		StopperAddress: stopperAddress}, nil
}

func HardStop(args *HardStopArgs) error {
	err := hardStop(args.TxHash, args.StopperAddress, args.VoteNonce, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("stop error 1: %v", err))
	}

	return nil
}

func hardStop(txHash, stopperAddress string, voteNonce int64, blockHeight int64) error {
	if stopperAddress != config.VoteSuperAddress {
		return errors.New("error 1: permission denied")
	}

	var (
		vote    Vote
		voteIdx int = -1
	)
	for idx := range VoteMemory {
		if VoteMemory[idx].Nonce == voteNonce {
			voteIdx = idx
			break
		}
	}

	if voteIdx == -1 {
		return errors.New("error 2: vote does not exist")
	}

	voteJson := VoteDB.Get(strconv.FormatInt(VoteMemory[voteIdx].Nonce, 10)).Value
	if voteJson != "" {
		err := json.Unmarshal([]byte(voteJson), &vote)
		if err != nil {
			return errors.New(fmt.Sprintf("error 3: %v", err))
		}
	}

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
	vote.EndTimestamp = contracts.String(timestamp)

	vote.Answers = VoteMemory[voteIdx].Answers

	jsonVote, err := json.Marshal(vote)

	if err != nil {
		return errors.New(fmt.Sprintf("error 4: %v", err))
	}

	err = contracts.AddEvent(config.VoteScAddress,
		*contracts.NewEvent("Hard stop", timestamp, blockHeight, txHash, stopperAddress,
			newEventHardStopTypeData(stopperAddress, blockHeight)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 5: %v", err))
	}

	VoteMemory = append(VoteMemory[:voteIdx], VoteMemory[voteIdx+1:]...)

	VoteDB.Put(strconv.FormatInt(VoteMemory[voteIdx].Nonce, 10), string(jsonVote))

	return nil
}
