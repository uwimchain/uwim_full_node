package validation

import (
	"encoding/json"
	"fmt"
	"node/apparel"
	"node/blockchain/contracts"
	"node/blockchain/contracts/delegate_con"
	"node/blockchain/contracts/delegate_con/delegate_validation"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"sort"
	"strings"

	"github.com/syndtr/goleveldb/leveldb/errors"
)

var (
	a = deep_actions.Address{}
)

func ValidateBlock(block deep_actions.Chain) error {
	if !memory.IsNodeProposer() {
		if block.Header.Proposer != "" {
			publicKey, _ := crypt.PublicKeyFromAddress(block.Header.Proposer)
			if !crypt.VerifySign(publicKey, []byte(block.Header.Proposer), block.Header.ProposerSignature) {
				return errors.New("Block validation error: proposer signature verify error.")
			}
		} else {
			return errors.New("Block validation error: proposer is null.")
		}
	}

	if config.BlockHeight != storage.BlockMemory.Height {
		return errors.New("Block validation error: blocks heights do not match.")
	}

	if config.BlockHeight == 0 {
		return errors.New("Block validation error: zero block not from the main node.")
	}

	if storage.GetPrevBlockHash() != block.Header.PrevHash {
		return errors.New("Block validation error: prev blocks hashes do not match.")
	}

	if storage.GetPrevBlockHash() != block.Header.PrevHash {
		return errors.New("Block validation error: prev blocks hashes do not match.")
	}

	if block.Header.Proposer != memory.Proposer {
		return errors.New("Block validation error: validator that sent the block for verification is not a proposer.")
	}

	if block.Header.TxCounter != int64(len(block.Txs)) {
		return errors.New("Block validation error: number of transactions in the block header does not match the number of transactions in the block body.")
	}

	if len(block.Txs) <= 0 {
		return errors.New("Block validation error: block haven`t transactions.")
	}

	if err := ValidateTxs(block.Txs); err != nil {
		return errors.New(fmt.Sprintf("Block transaction validation error: %v.", err))
	}

	for _, transaction := range block.Txs {
		if transaction.Type != 2 && !storage.FindTxInMemory(transaction.Nonce) {
			return errors.New("Block transaction validation error: transaction out of memory")
		}
	}

	return nil
}

func ValidateTxs(transactions []deep_actions.Tx) error {
	if addresses, err := fromAddressList(transactions); err != nil {
		return err
	} else {
		for _, transaction := range transactions {
			if fromAddressIdx := sort.Search(len(addresses), func(i int) bool { return addresses[i].Address >= transaction.From }); fromAddressIdx != len(addresses) {
				address := &addresses[fromAddressIdx]
				if err := validateTx(transaction, address); err != nil {
					return err
				} else {
					if tokenIdx := sort.Search(len(address.Balance), func(i int) bool { return address.Balance[i].TokenLabel >= transaction.TokenLabel }); tokenIdx != len(address.Balance) {
						token := &address.Balance[tokenIdx]
						token.Amount -= transaction.Amount + apparel.CalcTax(transaction.Amount*config.Tax)
					}
				}
			}
		}
	}

	return nil
}

func fromAddressList(transactions []deep_actions.Tx) ([]deep_actions.Address, error) {
	var allTransactionsAddresses []deep_actions.Address
	var transactionsAddresses []deep_actions.Address
	for _, transaction := range transactions {
		if row := a.GetAddress(transaction.From); row != "" {
			address := deep_actions.Address{}
			_ = json.Unmarshal([]byte(row), &address)
			allTransactionsAddresses = append(allTransactionsAddresses, address)
		} else {
			return nil, errors.New("senders address does not exist")
		}
	}

	for idx, address := range allTransactionsAddresses {
		if sort.Search(len(allTransactionsAddresses), func(i int) bool { return allTransactionsAddresses[i].Address >= address.Address }) == idx {
			transactionsAddresses = append(transactionsAddresses, address)
		}
	}

	return transactionsAddresses, nil
}

func validateTx(transaction deep_actions.Tx, address *deep_actions.Address) error {
	if transaction.Type != 5 {
		publicKey, err := crypt.PublicKeyFromAddress(transaction.From)
		if err != nil {
			return errors.New("signature verify error 1")
		}
		if !crypt.VerifySign(publicKey, []byte(transaction.From), transaction.Signature) {
			return errors.New("signature verify error 2")
		}
	}

	_, err := crypt.PublicKeyFromAddress(transaction.To)
	if err != nil {
		return errors.New("incorrect recipient address")
	}

	if transaction.Height == 0 {
		return errors.New("transaction block height is empty")
	}

	if transaction.Comment.Title != "undelegate_contract_transaction" && transaction.Amount <= 0 {
		return errors.New("zero or negative amount")
	}

	if !storage.CheckToken(transaction.TokenLabel) {
		return errors.New("token does not exist")
	}

	if transaction.From == config.GenesisAddress && transaction.Type != 2 {
		return errors.New("transaction from the genesis address")
	}

	switch transaction.Type {
	case 1:
		if !crypt.IsAddressUw(transaction.From) || transaction.From == config.GenesisAddress {
			return errors.New("this address haven`t permission for send transactions of this type")
		}

		switch transaction.Comment.Title {
		case "default_transaction":
			break
		case "delegate_contract_transaction":
			if transaction.To != config.DelegateScAddress {
				return errors.New("delegate transaction send not to this smart-contract address")
			}

			break
		case "undelegate_contract_transaction":
			if transaction.To != config.DelegateScAddress {
				return errors.New("undelegate transaction send not to this smart-contract address")
			}

			if transaction.Amount != 0 {
				return errors.New("undelegate transaction amount must be 0")
			}

			check := delegate_validation.UnDelegateValidate(transaction.From, transaction.Amount)
			switch check {
			case 1:
				return errors.New("low uwm-delegate balance")
			}

			for _, t := range storage.TransactionsMemory {
				if t.Type == 1 && t.From == transaction.From &&
					t.Comment.Title == "undelegate_contract_transaction" &&
					t.Nonce != transaction.Nonce {
					return errors.New("only one undelegate transaction can be send")
				}
			}

			break
		case "smart_contract_abandonment":
			{
				if transaction.Amount != 1 {
					return errors.New("amount of transaction of this type must be 1")
				}

				if transaction.TokenLabel != config.BaseToken {
					return errors.New("token label of transaction of this type must be uwm")
				}

				if transaction.To != config.GenesisAddress {
					return errors.New("transactions of this type must be send to genesis address")
				}

				if storage.CheckAddressScKeeping(transaction.From) {
					return errors.New("address haven`t smart-contract")
				}

				break
			}

		default:
			return errors.New("transaction type does not match the comment title 1:" + transaction.Comment.Title)
		}
		break
	case 2:
		if transaction.From != config.GenesisAddress {
			return errors.New("this address haven`t permission for send transactions of this type")
		}

		switch transaction.Comment.Title {
		case "reward_transaction":
			{
				publicKey, err := crypt.PublicKeyFromAddress(transaction.To)
				if err != nil {
					return errors.New("default reward transaction incorrect reward recipient")
				}
				toAddress := crypt.AddressFromPublicKey(metrics.NodePrefix, publicKey)

				if transaction.Amount != storage.CalculateReward(toAddress) {
					return errors.New("default reward transaction incorrect reward amount")
				} else {
					if transaction.TokenLabel != config.RewardTokenLabel {
						return errors.New("default reward transaction incorrect reward token")
					}
				}
			}
			break
		case "delegate_reward_transaction":
			{
				reward := storage.CalculateReward(config.DelegateScAddress)
				if transaction.Amount != reward {
					return errors.New(fmt.Sprintf("delegate reward transaction incorrect reward amount. My: %g, in transaction: %g", reward, transaction.Amount))
				} else {
					if transaction.TokenLabel != config.RewardTokenLabel {
						return errors.New(fmt.Sprintf("delegate reward transaction incorrect reward token. My: %v, in transaction: %s", config.RewardTokenLabel, transaction.TokenLabel))
					}
				}
			}
			break
		default:
			return errors.New("transaction type does not match the comment title 2")
		}
		break
	case 3:
		switch transaction.Comment.Title {
		case "create_token_transaction":
			t := deep_actions.Token{}
			err := json.Unmarshal(transaction.Comment.Data, &t)
			if err != nil {
				return errors.New("create token data error")
			}

			if t.Label == "" {
				return errors.New("token label is empty")
			}

			if int64(len(t.Label)) > config.MaxLabel {
				return errors.New("token label is greater than maximum")
			}

			if int64(len(t.Label)) < config.MinLabel {
				return errors.New("token label is less than the minimum")
			}

			if t.Name == "" {
				return errors.New("token name is empty")
			}

			if int64(len(t.Name)) > config.MaxName {
				return errors.New("token name is greater than maximum")
			}

			if t.Type != 0 {
				return errors.New("this type of token does not exist")
			}

			if t.Emission == 0 {
				return errors.New("token emission is empty")
			}

			if t.Emission > config.MaxEmission {
				return errors.New("token emission is greater than maximum")
			}
			if t.Emission < config.MinEmission {
				return errors.New("token emission is less than the minimum")
			}

			if t.CheckToken(t.Label) {
				return errors.New("token already exists")
			}

			if a.CheckAddressToken(t.Proposer) {
				return errors.New("this user have token")
			}

			balance := storage.GetBalance(t.Proposer)
			if balance != nil {
				for _, coin := range balance {
					if coin.TokenLabel == config.BaseToken {
						if t.Emission > 10000000 {
							if coin.Amount < config.NewTokenCost1 {
								return errors.New("low balance for create token")
							}
						} else if t.Emission > 10000000 {
							if coin.Amount < config.NewTokenCost2 {
								return errors.New("low balance for create token")
							}
						}
					}
				}
			} else {
				return errors.New("low balance for create token")
			}
			break
		case "rename_token_transaction":
			token := deep_actions.Token{}
			err := json.Unmarshal(transaction.Comment.Data, &token)
			if err != nil {
				return errors.New("rename token data error")
			}

			if token.Label == "" {
				return errors.New("token label is empty")
			}

			if token.Label == config.BaseToken {
				return errors.New("token label is \"uwm\"")
			}

			if token.Name == "" {
				return errors.New("token name is empty")
			}

			if int64(len(token.Name)) > config.MaxName {
				return errors.New("token name is greater than maximum")
			}

			if !token.CheckToken(token.Label) {
				return errors.New("this token does not exist`s")
			}

			if !a.CheckAddressToken(transaction.From) {
				return errors.New("this user haven`t token")
			}

			balance := storage.GetBalance(transaction.From)
			if balance != nil {
				for _, coin := range balance {
					if coin.TokenLabel == config.BaseToken {
						if coin.Amount < config.RenameTokenCost {
							return errors.New(fmt.Sprintf("low balance %s for rename token. Balance: %g", transaction.From, coin.Amount))
						}
					}
				}
			} else {
				return errors.New(fmt.Sprintf("low balance %s for rename token", transaction.From))
			}
			break
		case "change_token_standard_transaction":
			t := deep_actions.Token{}
			err := json.Unmarshal(transaction.Comment.Data, &t)
			if err != nil {
				return errors.New("change token standard data error")
			}

			if !apparel.SearchInArray([]int64{0, 1, 2, 3, 4}, t.Standard) {
				return errors.New("invalid token standard 1")
			}

			row := storage.GetToken(t.Label)
			if row == "" {
				return errors.New("invalid token standard 2")
			}
			token := deep_actions.Token{}
			err = json.Unmarshal([]byte(row), &token)
			if err != nil {
				return errors.New("invalid token standard 3")
			}

			if !storage.CheckToken(token.Label) {
				return errors.New("invalid token standard 4")
			}

			if t.Standard == token.Standard {
				return errors.New("invalid token standard 5")
			}

			if token.Standard == 0 && !apparel.SearchInArray([]int64{1, 2}, t.Standard) {
				return errors.New("invalid token standard 6")
			}

			if apparel.SearchInArray([]int64{1, 2}, token.Standard) && !apparel.SearchInArray([]int64{1, 2, 3, 4}, t.Standard) {
				return errors.New("invalid token standard 7")
			}

			if apparel.SearchInArray([]int64{3, 4}, token.Standard) && !apparel.SearchInArray([]int64{3, 4}, t.Standard) {
				return errors.New("invalid token standard 8")
			}

			break
		case "fill_token_card_transaction":
			tokenCard := deep_actions.PersonalTokenCard{}
			err := json.Unmarshal(transaction.Comment.Data, &tokenCard)
			if err != nil {
				return errors.New("fill token card data error")
			}

			if tokenCard.Hashtags != "" {
				hashtags := strings.Split(strings.TrimSpace(tokenCard.Hashtags), "#")
				if hashtags != nil && len(hashtags)-1 < 3 {
					return errors.New("invalid token card data 16")
				}

				if hashtags != nil && len(hashtags)-1 > 10 {
					return errors.New("invalid token card data 17")
				}

			}

			break
		case "fill_token_standard_card_transaction":
			token := storage.GetAddressToken(transaction.From)
			switch token.Standard {
			case 0:
				if check := validate0standard(string(transaction.Comment.Data)); check != nil {
					return check
				}
				break
			case 1:
				if check := validate1standard(string(transaction.Comment.Data)); check != nil {
					return check
				}
				break
			case 2:
				if check := validate2standard(string(transaction.Comment.Data)); check != nil {
					return check
				}
				break
			case 3:
				if check := validate3standard(string(transaction.Comment.Data)); check != nil {
					return check
				}
				break
			case 4:
				if check := validate4standard(string(transaction.Comment.Data)); check != nil {
					return check
				}
				break
			}
			break
		default:
			return errors.New("transaction type does not match the comment title 3")
		}
		break
	case 5:
		switch transaction.Comment.Title {
		case "undelegate_contract_transaction":
			{
				check := delegate_validation.UnDelegateValidate(transaction.To, transaction.Amount)
				switch check {
				case 1:
					return errors.New("low uwm-delegate balance")
				case 2:
					return errors.New("delegate smart-contract address haven`t coins for transaction")
				}

				contractData := contracts.ContractCommentData{}
				err := json.Unmarshal(transaction.Comment.Data, &contractData)
				if err != nil {
					return errors.New("contract data error")
				} else {
					publicKey, _ := crypt.PublicKeyFromAddress(contractData.NodeAddress)
					if !crypt.VerifySign(publicKey, []byte(transaction.To), transaction.Signature) {
						return errors.New("signature verify error")
					}
				}

				if err := delegate_con.UnDelegateValidate(transaction.To, transaction.Amount, contractData.CheckSum); err != nil {
					return errors.New("checksum verify error")
				}
			}
			break
		default:
			return errors.New("transaction type does not match the comment title 5")
		}
		break
	default:
		return errors.New("transaction type is empty")
	}

	if address.Address == transaction.From {
		sendToken := deep_actions.Balance{}
		taxToken := deep_actions.Balance{}

		for _, token := range address.Balance {
			if token.TokenLabel == transaction.TokenLabel {
				sendToken = token
			}

			if token.TokenLabel == config.BaseToken {
				taxToken = token
			}
		}

		if sendToken.Amount < transaction.Amount {
			return errors.New("low balance")
		}

		if taxToken.Amount < apparel.CalcTax(transaction.Amount*config.TaxConversion*config.Tax) {
			return errors.New("low balance")
		}

	}

	return nil
}

func validate0standard(data string) error {
	return nil
}

func validate1standard(data string) error {
	tokenStandardCard := deep_actions.ThxStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return errors.New("invalid token standard card data 111")
	}

	return nil
}

func validate2standard(data string) error {
	tokenStandardCard := deep_actions.DropStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return errors.New("invalid token standard card data 121")
	}

	return nil
}

func validate3standard(data string) error {
	tokenStandardCard := deep_actions.StartUpStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return errors.New("invalid token standard card data 131")
	}

	return nil
}

func validate4standard(data string) error {
	tokenStandardCard := deep_actions.BusinessStandardCardData{}
	err := json.Unmarshal([]byte(data), &tokenStandardCard)
	if err != nil {
		return errors.New("invalid token standard card data 141")
	}

	return nil
}
