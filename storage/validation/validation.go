package validation

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"sort"
)

var (
	a = deep_actions.Address{}
)

// Функция для валидации блока
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

// Функция для валидации полученного списка транзакций
func ValidateTxs(transactions []deep_actions.Tx) error {
	if addresses, err := fromAddressList(transactions); err != nil {
		return err
	} else {
		for _, transaction := range transactions {
			if fromAddressIdx := sort.Search(len(addresses),
				func(i int) bool { return addresses[i].Address >= transaction.From }); fromAddressIdx != len(addresses) {
				address := &addresses[fromAddressIdx]
				if err := ValidateTx(transaction, address); err != nil {
					return err
				} else {
					if tokenIdx := sort.Search(len(address.Balance),
						func(i int) bool { return address.Balance[i].TokenLabel >= transaction.TokenLabel });
						tokenIdx != len(address.Balance) {
						token := &address.Balance[tokenIdx]
						token.Amount -= transaction.Amount + apparel.CalcTax(transaction.Amount)
					}
				}
			}
		}
	}

	return nil
}

//Apparel
// вспомогательная функция для получения все адресов отправителей из введённого списка транзакций для произведения дальнейшей валидации
func fromAddressList(transactions []deep_actions.Tx) ([]deep_actions.Address, error) {
	var allTransactionsAddresses []deep_actions.Address
	var transactionsAddresses []deep_actions.Address
	for _, transaction := range transactions {
		if transaction.From != config.VoteSuperAddress {
			if row := a.GetAddress(transaction.From); row != "" {
				address := deep_actions.Address{}
				_ = json.Unmarshal([]byte(row), &address)
				allTransactionsAddresses = append(allTransactionsAddresses, address)
			} else {
				return nil, errors.New("senders address does not exist")
			}
		}
	}

	for idx, address := range allTransactionsAddresses {
		if sort.Search(len(allTransactionsAddresses),
			func(i int) bool { return allTransactionsAddresses[i].Address >= address.Address }) == idx {
			transactionsAddresses = append(transactionsAddresses, address)
		}
	}

	return transactionsAddresses, nil
}

// вспомогательная функция для валидации транзакции
func ValidateTx(transaction deep_actions.Tx, address *deep_actions.Address) error {
	if transaction.Type != 5 {
		publicKey, err := crypt.PublicKeyFromAddress(transaction.From)
		if err != nil {
			return errors.New("signature verify error 1")
		}
		if !crypt.VerifySign(publicKey, []byte(transaction.From), transaction.Signature) {
			if transaction.Comment.Title == "default_transaction" && transaction.Comment.Data != nil {
				buyTokenSign := deep_actions.BuyTokenSign{}
				err := json.Unmarshal(transaction.Comment.Data, &buyTokenSign)
				if err != nil {
					return errors.New("signature verify error 2")
				}

				if !memory.IsNodeValidator(buyTokenSign.NodeAddress) {
					return errors.New("signature verify error 3")
				}

				nodePublicKey, err := crypt.PublicKeyFromAddress(buyTokenSign.NodeAddress)
				if err != nil {
					return errors.New("signature verify error 4")
				}

				if !crypt.VerifySign(nodePublicKey, []byte(buyTokenSign.NodeAddress), transaction.Signature) {
					return errors.New("signature verify error 5")
				}
			}
		}
	}

	_, err := crypt.PublicKeyFromAddress(transaction.To)
	if err != nil {
		return errors.New("incorrect recipient address")
	}

	if transaction.Height == 0 {
		return errors.New("transaction block height is empty")
	}

	if transaction.Comment.Title != "undelegate_contract_transaction" &&
		transaction.Comment.Title != "my_token_contract_confirmation_transaction" &&
		transaction.Comment.Title != "trade_token_contract_get_com_transaction" &&
		transaction.Comment.Title != "trade_token_contract_get_liq_transaction" &&
		transaction.Comment.Title != "holder_contract_get_transaction" &&
		transaction.Comment.Title != "vote_contract_start_transaction" &&
		transaction.Comment.Title != "vote_contract_hard_stop_transaction" &&
		transaction.Amount <= 0 {
		return errors.New("zero or negative amount")
	}

	if transaction.TokenLabel != config.BaseToken && !storage.CheckToken(transaction.TokenLabel) {
		return errors.New(fmt.Sprintf("token \"%s\" does not exist", transaction.TokenLabel))
	}

	// Валидация на случай, если транзакция отправлена с Genesis адреса, но не является наградой
	if transaction.From == config.GenesisAddress && transaction.Type != 2 {
		return errors.New("transaction from the genesis address")
	}

	// Валидация комментария к транзакции в зависимости от её типа
	// и проверка соотвествия типа транзакции заголовку её комментария
	switch transaction.Type {
	case 1:
		if err := validateTransactionType1(transaction); err != nil {
			return err
		}
		break
	case 2:
		if err := validateTransactionType2(transaction); err != nil {
			return err
		}
		break
	case 3:
		if err := validateTransactionType3(transaction); err != nil {
			return err
		}
		break
	case 5:
		if err := validateTransactionType5(transaction); err != nil {
			return err
		}
		break
	default:
		return errors.New("transaction type is empty")
	}

	// Валидация баланса пользователя в зависимости от выбранного токена транзакции
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
			return errors.New(fmt.Sprintf("low balance %g %s %s %s", sendToken.Amount, sendToken.TokenLabel, transaction.TokenLabel, transaction.From))
		}

		switch transaction.Comment.Title {
		case "undelegate_contract_transaction":
		case "refund_transaction":
		case "my_token_contract_confirmation_transaction":
		case "my_token_contract_get_percent_transaction":
		case "business_token_contract_get_percent_transaction":
		case "trade_token_contract_get_liq_transaction":
		case "trade_token_contract_get_com_transaction":
		case "trade_token_contract_fill_config_transaction":
		case "vote_contract_start_transaction":
		case "vote_contract_hard_stop_transaction":
			if transaction.Tax != 0 {
				return errors.New("invalid transaction tax amount")
			}
			break
		case "holder_contract_get_transaction":
		case "holder_contract_add_transaction":
			if taxToken.Amount < transaction.Amount {
				return errors.New(fmt.Sprintf("low balance for tax. need %g there is %g for transaction with type %s", transaction.Tax, taxToken.Amount, transaction.Comment.Title))
			}
			break
		default:
			if taxToken.Amount < apparel.CalcTax(transaction.Amount) {
				return errors.New(fmt.Sprintf("low balance for tax. need %g there is %g for transaction with type %s", transaction.Tax, taxToken.Amount, transaction.Comment.Title))
			}
			break
		}
	}

	return nil
}
