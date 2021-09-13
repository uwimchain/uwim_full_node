package validation

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/blockchain/contracts"
	"node/blockchain/contracts/delegate_con"
	"node/blockchain/contracts/delegate_con/delegate_validation"
	"node/crypt"
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

			contractData := contracts.ContractCommentData{}
			err := json.Unmarshal(t.Comment.Data, &contractData)
			if err != nil {
				return errors.New("contract data error")
			} else {
				publicKey, _ := crypt.PublicKeyFromAddress(contractData.NodeAddress)
				if !crypt.VerifySign(publicKey, []byte(t.To), t.Signature) {
					return errors.New("signature verify error")
				}
			}

			if err := delegate_con.UnDelegateValidate(t.To, t.Amount, contractData.CheckSum); err != nil {
				return errors.New("checksum verify error")
			}

			break
		}
	default:
		return errors.New("transaction type does not match the comment title 5")
	}

	return nil
}
