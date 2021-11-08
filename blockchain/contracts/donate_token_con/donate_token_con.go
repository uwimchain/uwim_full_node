package donate_token_con

import (
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/blockchain/contracts"
)

func ChangeStandard(scAddress string) error {
	scBalance := contracts.GetBalance(scAddress)
	if scBalance == nil {
		return nil
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return errors.New("Donate token contract error 1: token does not exist")
	}

	for _, i := range scBalance {
		_ = contracts.RefundTransaction(scAddress, token.Proposer, i.Amount, i.TokenLabel)
	}

	return nil
}
