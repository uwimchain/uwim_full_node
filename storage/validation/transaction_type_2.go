package validation

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
)

func validateTransactionType2(t deep_actions.Tx) error {
	if t.From != config.GenesisAddress {
		return errors.New("this address haven`t permission for send transactions of this type")
	}

	switch t.Comment.Title {
	case "reward_transaction":

		publicKey, err := crypt.PublicKeyFromAddress(t.To)
		if err != nil {
			return errors.New("default reward transaction incorrect reward recipient")
		}
		toAddress := crypt.AddressFromPublicKey(metrics.NodePrefix, publicKey)

		reward, _ := apparel.Round(storage.CalculateReward(toAddress))
		if t.Amount != reward {
			return errors.New("default reward transaction incorrect reward amount")
		} else {
			if t.TokenLabel != config.RewardTokenLabel {
				return errors.New("default reward transaction incorrect reward token")
			}
		}
		break
	case "delegate_reward_transaction":

		reward, _ := apparel.Round(storage.CalculateReward(config.DelegateScAddress))
		if t.Amount != reward {
			log.Println(storage.GetBalanceForToken(config.DelegateScAddress, config.BaseToken))
			return errors.New(
				fmt.Sprintf("delegate reward transaction incorrect reward amount. My: %g, in transaction: %g",
					reward, t.Amount))
		} else {
			if t.TokenLabel != config.RewardTokenLabel {
				return errors.New(fmt.Sprintf(
					"delegate reward transaction incorrect reward token. My: %v, in transaction: %s",
					config.RewardTokenLabel, t.TokenLabel))
			}
		}
		break
	default:
		return errors.New("transaction type does not match the comment title 2")
	}

	return nil
}
