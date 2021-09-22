package business_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"strconv"
)

type GetPercentArgs struct {
	ScAddress   string  `json:"sc_address"`
	UwAddress   string  `json:"uw_address"`
	Amount      float64 `json:"amount"`
	TokenLabel  string  `json:"token_label"`
	TxHash      string  `json:"tx_hash"`
	BlockHeight int64   `json:"block_height"`
}

func NewGetPercentArgs(scAddress string, uwAddress, txHash string, data interface{}, blockHeight int64) (*GetPercentArgs, error) {
	scAddress, uwAddress, txHash = apparel.TrimToLower(scAddress), apparel.TrimToLower(uwAddress), apparel.TrimToLower(txHash)

	if !crypt.IsAddressSmartContract(scAddress) {
		return nil, errors.New(fmt.Sprintf("error 1: this address \"%s\" is not a smart-contract address", scAddress))
	}

	if !crypt.IsAddressUw(uwAddress) {
		return nil, errors.New(fmt.Sprintf("error 2: this address \"%s\" is not a uwim address", scAddress))
	}

	if txHash == "" || storage.GetTxForHash(txHash) == "" {
		return nil, errors.New(fmt.Sprintf("error 3: transaction width this hash \"%s\" does not exist", txHash))
	}

	if data == nil {
		return nil, errors.New("error 4: apparel data for business contract take percent transaction in null")
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error 5: %v", err))
	}
	dataArr := make(map[string]interface{})
	err = json.Unmarshal(dataJson, &dataArr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error 6: %v", err))
	}

	amount, _ := apparel.Round(apparel.ConvertInterfaceToFloat64(dataArr["amount"]))
	if amount <= 0 {
		return nil, errors.New("error 7: zero or negative amount")
	}

	tokenLabel := apparel.TrimToLower(apparel.ConvertInterfaceToString(dataArr["token_label"]))

	return &GetPercentArgs{ScAddress: scAddress, UwAddress: uwAddress, Amount: amount, TokenLabel: tokenLabel, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func GetPercent(args *GetPercentArgs) error {
	err := getPercent(args.ScAddress, args.UwAddress, args.TokenLabel, args.TxHash, args.Amount, args.BlockHeight)
	if err != nil {
		return errors.New(fmt.Sprintf("error 1: %v", err))
	}

	return nil
}

func getPercent(scAddress, uwAddress, tokenLabel, txHash string, amount float64, blockHeight int64) error {

	if !crypt.IsAddressSmartContract(scAddress) {
		return errors.New(fmt.Sprintf("error 1: this address \"%s\" is not a smart-contract address", scAddress))
	}

	if !crypt.IsAddressUw(uwAddress) {
		return errors.New(fmt.Sprintf("error 2: this address \"%s\" is not a uwim address", scAddress))
	}

	if amount <= 0 {
		return errors.New("error 3: zero or negative amount")
	}

	var scAddressPartners []Partner
	scAddressPartnersJson := ContractsDB.Get(scAddress).Value
	if scAddressPartnersJson != "" {
		err := json.Unmarshal([]byte(scAddressPartnersJson), &scAddressPartners)
		if err != nil {
			return errors.New(fmt.Sprintf("error 4: %v", err))
		}
	}

	timestamp := apparel.TimestampUnix()
	partnerExist := false
	for idx, i := range scAddressPartners {
		if i.Address == uwAddress {
			if i.Balance == nil {
				return errors.New(fmt.Sprintf("error 5: balance of \"%s\" is null", uwAddress))
			}

			tokenExist := false
			for jdx, j := range i.Balance {
				if j.TokenLabel == tokenLabel {
					if j.Amount < amount {
						return errors.New("error 6: low balance")
					}

					scAddressPartners[idx].Balance[jdx].Amount -= amount
					scAddressPartners[idx].Balance[jdx].UpdateTime = strconv.FormatInt(timestamp, 10)

					tokenExist = true
					break
				}
			}
			if !tokenExist {
				return errors.New(fmt.Sprintf("error 7: token \"%s\" does not exist on partner \"%s\" balance", tokenLabel, uwAddress))
			}
			partnerExist = true

			break
		}
	}

	if !partnerExist {
		return errors.New(fmt.Sprintf("error 8: partner \"%s\" does not exist on pertners list of \"%s\"", uwAddress, scAddress))
	}

	txTax := apparel.CalcTax(amount)

	scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, tokenLabel)
	scAddressBalanceForTaxToken := contracts.GetBalanceForToken(scAddress, config.BaseToken)

	if scAddressBalanceForToken.TokenLabel != scAddressBalanceForTaxToken.TokenLabel {
		if scAddressBalanceForToken.Amount < amount || scAddressBalanceForToken.Amount-amount < 1 {
			return errors.New(fmt.Sprintf("error 9: smart-contract has low balance for token \"%s\"", tokenLabel))
		}

		if scAddressBalanceForTaxToken.Amount < amount || scAddressBalanceForTaxToken.Amount-txTax < 1 {
			return errors.New(fmt.Sprintf("error 10: smart-contract has low balance for token \"%s\"", tokenLabel))
		}
	} else {
		if scAddressBalanceForToken.Amount < amount+txTax || scAddressBalanceForToken.Amount-amount-txTax < 1 {
			return errors.New(fmt.Sprintf("error 11: smart-contract has low balance for token \"%s\"", tokenLabel))
		}
	}

	txCommentSign, _ := json.Marshal(contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	))

	tx := contracts.NewTx(
		5,
		apparel.GetNonce(strconv.FormatInt(timestamp, 10)),
		"",
		config.BlockHeight,
		scAddress,
		uwAddress,
		amount,
		tokenLabel,
		strconv.FormatInt(timestamp, 10),
		txTax,
		nil,
		*contracts.NewComment(
			"default_transaction",
			txCommentSign,
		),
	)

	jsonString, _ := json.Marshal(contracts.Tx{
		Type:       tx.Type,
		Nonce:      tx.Nonce,
		From:       tx.From,
		To:         tx.To,
		Amount:     tx.Amount,
		TokenLabel: tx.TokenLabel,
		Comment:    tx.Comment,
	})
	tx.Signature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)

	err := contracts.AddEvent(scAddress, *contracts.NewEvent("Get percent", timestamp, blockHeight, txHash, uwAddress, newEventGetPercentTypeData(amount, tokenLabel)), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 13: %v", err))
	}

	jsonScAddressPartners, err := json.Marshal(scAddressPartners)
	if err != nil {
		return errors.New(fmt.Sprintf("error 14: %v", err))
	}
	ContractsDB.Put(scAddress, string(jsonScAddressPartners))

	if memory.IsNodeProposer() {
		contracts.SendTx(*tx)
		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *tx)
	}

	return nil
}
