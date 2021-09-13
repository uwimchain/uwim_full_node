package api

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/blockchain/contracts/vote_con"
)

type VoteContractGetVotesArgs struct {
}

func (api *Api) VoteContractGetVotes(args *VoteContractHardStopArgs, result *string) error {
	votes := vote_con.GetVotes()
	voteJson, err := json.Marshal(votes)
	if err != nil {
		return errors.New(fmt.Sprintf("Vote contract get votes error 1: %v", err))
	}

	*result = string(voteJson)

	return nil
}
