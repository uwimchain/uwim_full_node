package validation

import (
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/blockchain/contracts/delegate_con/delegate_validation"
	"node/storage/deep_actions"
)

func validateTransactionType5(t deep_actions.Tx) error {
	switch t.Comment.Title {
	case "default_transaction":
		// pass
		break
	case "refund_transaction":
		// pass
		break
	case "undelegate_contract_transaction":
		{
			check := delegate_validation.UnDelegateValidate(t.To, t.Amount)
			switch check {
			case 1:
				return errors.New("low uwm-delegate balance")
			case 2:
				return errors.New("delegate smart-contract address haven`t coins for transaction")
			}

			break
		}
	default:
		return errors.New("transaction type does not match the comment title 5")
	}

	return nil
}
