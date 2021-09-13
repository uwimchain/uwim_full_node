package donate_token_con

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/memory"
	"strconv"
)

type BuyArgs struct {
	ScAddress   string  `json:"sc_address"`
	UwAddress   string  `json:"uw_address"`
	TokenLabel  string  `json:"token_label"`
	Amount      float64 `json:"amount"`
	TxHash      string  `json:"tx_hash"`
	BlockHeight int64   `json:"block_height"`
}

func NewBuyArgs(scAddress string, uwAddress string, tokenLabel string, amount float64, txHash string, blockHeight int64) (*BuyArgs, error) {
	amount,_ = apparel.Round(amount)
	return &BuyArgs{ScAddress: scAddress, UwAddress: uwAddress, TokenLabel: tokenLabel, Amount: amount, TxHash: txHash, BlockHeight: blockHeight}, nil
}

func Buy(args *BuyArgs) error {
	err := buy(args.ScAddress, args.UwAddress, args.TokenLabel, args.TxHash, args.Amount, args.BlockHeight)
	if err != nil {
		refundError := contracts.RefundTransaction(args.ScAddress, args.UwAddress, args.Amount, args.TokenLabel)
		if refundError != nil {
			log.Println(fmt.Sprintf("Refund transaction %v", refundError))
		}
		return errors.New(fmt.Sprintf("error 1: buy %v", err))
	}

	return nil
}

func buy(scAddress, uwAddress, tokenLabel, txHash string, amount float64, blockHeight int64) error {
	timestamp := apparel.TimestampUnix()

	if !crypt.IsAddressSmartContract(scAddress) {
		return errors.New("error 1: invalid smart-contract address")
	}

	if !crypt.IsAddressUw(uwAddress) {
		return errors.New("error 2: invalid uw address")
	}

	if amount <= 0 {
		return errors.New("error 3: zero or negative amount")
	}

	scAddressToken := contracts.GetTokenInfoForScAddress(scAddress)
	if scAddressToken.Id == 0 {
		return errors.New("error 4: token does not exist")
	}

	if scAddressToken.Label != tokenLabel {
		return errors.New("error 5: token label not a smart-contract token label")
	}

	if scAddressToken.Proposer == uwAddress {
		return errors.New("error 6: sender address is a token proposer")
	}

	scAddressBalanceForToken := contracts.GetBalanceForToken(scAddress, scAddressToken.Label)
	if scAddressBalanceForToken.Amount <= 0 {
		return errors.New(fmt.Sprintf("error 7: smart-contract balance for token %s", scAddressToken.Label))
	}

	scAddressTokenStandardCard := contracts.DonateStandardCardData
	err := json.Unmarshal([]byte(scAddressToken.StandardCard), &scAddressTokenStandardCard)
	if err != nil {
		return errors.New(fmt.Sprintf("error 8: %v", err))
	}

	txAmount := scAddressTokenStandardCard.Conversion * amount

	if scAddressBalanceForToken.Amount < txAmount {
		return errors.New(fmt.Sprintf("erorr 9: smart-contract balance for token %s", scAddressToken.Label))
	}

	scAddressBalanceForTokenUwm := contracts.GetBalanceForToken(scAddress, config.BaseToken)
	txTax := apparel.CalcTax(txAmount)
	if scAddressBalanceForTokenUwm.Amount < txTax {
		return errors.New("error 10: smart-contract balance for token uwm")
	}

	txCommentSign, _ := json.Marshal(contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	))

	transaction := contracts.NewTx(
		5,
		apparel.GetNonce(strconv.FormatInt(timestamp, 10)),
		"",
		config.BlockHeight,
		scAddress,
		uwAddress,
		txAmount,
		scAddressToken.Label,
		strconv.FormatInt(timestamp, 10),
		txTax,
		crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
		*contracts.NewComment(
			"default_transaction",
			txCommentSign,
		),
	)

	jsonString, err := json.Marshal(transaction)
	if err != nil {
		return errors.New(fmt.Sprintf("error 11: %v", err))
	}

	err = contracts.AddEvent(scAddress, *contracts.NewEvent("Buy", timestamp, blockHeight, txHash, uwAddress, nil), EventDB, ConfigDB)
	if err != nil {
		return errors.New(fmt.Sprintf("error 12: %v", err))
	}

	if memory.IsNodeProposer() {
		contracts.SendTx(jsonString)
		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
	}

	return nil
}
