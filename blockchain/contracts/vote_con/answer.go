package vote_con

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/crypt"
)

type AnswerArgs struct {
	Address             string `json:"address"`
	Signature           []byte `json:"signature"`
	TxHash              string `json:"tx_hash"`
	PossibleAnswerNonce string  `json:"possible_answer_nonce"`
	VoteNonce           string  `json:"vote_nonce"`
	BlockHeight         int64  `json:"block_height"`
}

func NewAnswerArgs(address string, signature []byte, txHash string, possibleAnswerNonce string, voteNonce string,
	blockHeight int64) (*AnswerArgs, error) {
	return &AnswerArgs{Address: address, Signature: signature, TxHash: txHash, PossibleAnswerNonce: possibleAnswerNonce,
		VoteNonce: voteNonce, BlockHeight: blockHeight}, nil
}

func Answer(args *AnswerArgs) error {
	err := answer(args.Address, args.TxHash, args.Signature, args.PossibleAnswerNonce, args.VoteNonce, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("answer error 1: %v", err))
	}
	return nil
}

func answer(address, txHash string, signature []byte, possibleAnswerNonce, voteNonce string, blockHeight int64) error {
	if !crypt.IsAddressUw(address) && !crypt.IsAddressSmartContract(address) && !crypt.IsAddressNode(address) {
		return errors.New("error 1: empty or incorrect address")
	}

	publicKey, err := crypt.PublicKeyFromAddress(address)
	if err != nil {
		return errors.New(fmt.Sprintf("error 2: %v", err))
	}

	if !crypt.VerifySign(publicKey, []byte(address), signature) {
		return errors.New("error 3: verify signature")
	}

	if VoteMemory == nil {
		return errors.New("error 4: vote does not exist")
	}

	var (
		voteIdx           int = -1
		possibleAnswerIdx int = -1
	)
	for idx, i := range VoteMemory {
		if i.Nonce == voteNonce {
			voteIdx = idx

			if i.Answers != nil {
				for _, j := range i.Answers {
					if j.Address == address {
						return errors.New(fmt.Sprintf("error 5: this address \"%s\" exist`s in vote \"%v\"", address, VoteMemory[idx].Nonce))
					}
				}
			}

			if i.AnswerOptions == nil {
				return errors.New(fmt.Sprintf("error 6: empty answer options for this vote \"%d\"", VoteMemory[idx].Nonce))
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
		return errors.New(fmt.Sprintf("error 7: vote with this nonce \"%d\" does not exist", voteNonce))
	}

	if possibleAnswerIdx < 0 {
		return errors.New(fmt.Sprintf("error 8: answer option with this nonce \"%d\" does not exist in answer options list of vote with nonce \"%d\"", possibleAnswerNonce, voteNonce))
	}

	VoteMemory[voteIdx].Answers = append(VoteMemory[voteIdx].Answers, *NewAddressAnswer(address, signature, txHash, blockHeight, possibleAnswerNonce, voteNonce))

	return nil
}
