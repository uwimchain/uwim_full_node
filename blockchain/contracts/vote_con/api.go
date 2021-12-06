package vote_con

import (
	"encoding/json"
	"log"
	"node/config"
	"node/crypt"
)

func GetVotesJson() interface{} {
	return VoteMemory
}

func GetVotes() []Vote {
	return VoteMemory
}

//func GetVoteForNonce(voteNonce string) Vote {
func GetVoteForNonce(voteNonce int64) Vote {
	var vote Vote
	for _, i := range VoteMemory {
		if i.Nonce == voteNonce {
			return i
		}
	}

	return vote
}

//func GetVoteForNonceJson(voteNonce string) interface{} {
func GetVoteForNonceJson(voteNonce int64) interface{} {
	for _, i := range VoteMemory {
		if i.Nonce == voteNonce {
			return i
		}
	}

	return nil
}

func ValidateStart(title, description, starterAddress string, answerOptions interface{}, endBlockHeight int64) int64 {
	if starterAddress != config.VoteSuperAddress {
		return 811
	}

	if title == "" {
		return 812
	}

	if description == "" {
		return 813
	}

	if answerOptions == nil {
		return 814
	}

	answerOptionsJson, err := json.Marshal(answerOptions)
	if err != nil {
		log.Println("validate start error 1: ", err)
		return 815
	}

	var answerOptionsData []AnswerOption
	if answerOptionsJson != nil {
		err = json.Unmarshal(answerOptionsJson, &answerOptionsData)
		if err != nil {
			log.Println("validate start error 2: ", err)
			return 816
		}
	}

	if len(answerOptionsData) > config.MaxVoteAnswerOptions {
		return 817
	}

	if endBlockHeight <= config.BlockHeight {
		return 818
	}

	if VoteMemory != nil {
		if len(VoteMemory) >= config.MaxVoteMemory {
			return 819
		}
	}

	return 0
}

//func ValidateHardStop(stopperAddress string, voteNonce string) int64 {
func ValidateHardStop(stopperAddress string, voteNonce int64) int64 {
	if stopperAddress != config.VoteSuperAddress {
		return 821
	}

	var voteIdx int = -1
	for idx := range VoteMemory {
		if VoteMemory[idx].Nonce == voteNonce {
			voteIdx = idx
			break
		}
	}

	if voteIdx == -1 {
		return 822
	}

	return 0
}

//func ValidateAnswer(address string, voteNonce, possibleAnswerNonce string) int64 {
func ValidateAnswer(address string, voteNonce int64, possibleAnswerNonce string) int64 {
	if !crypt.IsAddressUw(address) && !crypt.IsAddressSmartContract(address) && !crypt.IsAddressNode(address) {
		return 831
	}

	if VoteMemory == nil {
		return 832
	}

	var (
		voteIdx           int = -1
		possibleAnswerIdx int = -1
	)

	log.Println("GG", voteNonce)
	for idx, i := range VoteMemory {
		log.Println(VoteMemory[idx].Nonce, i.Nonce)
		if i.Nonce == voteNonce {
			log.Println("FF")
			voteIdx = idx

			if i.Answers != nil {
				for _, j := range i.Answers {
					if j.Address == address {
						return 833
					}
				}
			}

			if i.AnswerOptions == nil {
				return 834
			}

			for jdx, j := range i.AnswerOptions {
				if j.PossibleAnswerNonce == possibleAnswerNonce {
					possibleAnswerIdx = jdx
					break
				}
			}

			break
		}
	}

	if voteIdx < 0 {
		return 835
	}

	if possibleAnswerIdx < 0 {
		return 836
	}

	return 0
}
