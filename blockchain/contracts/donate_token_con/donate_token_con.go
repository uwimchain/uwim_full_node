package donate_token_con

/*type DonateTokenConSmartArgs struct {
	ScAddress     string  `json:"sc_address"`
	SenderAddress string  `json:"sender_address"`
	Amount        float64 `json:"amount"`
}*/

/*func Smart(args DonateTokenConSmartArgs) {
	timestamp := apparel.TimestampUnix()
	timestampD := strconv.FormatInt(timestamp, 10)
	scAddress := args.ScAddress

	if !crypt.IsAddressSmartContract(scAddress) || scAddress == "" {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "sc address is null or not sc address",
		})

		return
	}

	if args.Amount <= 0 {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "zero or negative amount",
		})

		return
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "token does not exists",
		})

		refundTx(args)

		return
	}

	if token.Proposer == args.SenderAddress {
		return
	}

	scTokenBalance := getScTokenBalance(scAddress, token.Label)
	if scTokenBalance <= 0 {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "sc address have low balance",
		})

		refundTx(args)

		return
	}

	tokenStandardCard := contracts.DonateStandardCardData
	err := json.Unmarshal([]byte(token.StandardCard), &tokenStandardCard)
	if err != nil {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     fmt.Sprintf("%v", err),
		})

		refundTx(args)

		return
	}

	amount := tokenStandardCard.Conversion * args.Amount

	if scTokenBalance < amount {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "low balance",
		})

		refundTx(args)

		return
	}

	scBaseTokenBalance := getScTokenBalance(scAddress, config.BaseToken)
	tax := apparel.CalcTax(amount * config.Tax)
	if scBaseTokenBalance < tax {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "low base token balance",
		})

		refundTx(args)

		return
	}

	nonce := apparel.GetNonce(timestampD)

	commentSign, err := json.Marshal(contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	))
	if err != nil {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     fmt.Sprintf("%v", err),
		})

		refundTx(args)

		return
	}

	transaction := contracts.NewTx(
		5,
		nonce,
		"",
		config.BlockHeight,
		scAddress,
		args.SenderAddress,
		amount,
		token.Label,
		timestampD,
		tax,
		crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
		*contracts.NewComment(
			"default_transaction",
			commentSign,
		),
	)

	jsonString, err := json.Marshal(transaction)
	if err != nil {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     fmt.Sprintf("%v", err),
		})

		refundTx(args)

		return
	}

	if memory.IsNodeProposer() {
		contracts.SendTx(jsonString)
		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
	}

	newTx(scAddress, jsonString)
	//newTxs(nonce, jsonString)
}*/

/*func newTx(scAddress string, jsonTransaction []byte) {
	timestamp := apparel.TimestampUnix()
	timestampD := strconv.FormatInt(timestamp, 10)
	var transactions []contracts.Tx
	transaction := contracts.Tx{}
	jsonTransactions := TxDB.Get(scAddress).Value
	err := json.Unmarshal([]byte(jsonTransactions), &transactions)
	if err != nil {
		log(scAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     fmt.Sprintf("%v", err),
		})

		return
	}

	err = json.Unmarshal(jsonTransaction, &transaction)
	if err != nil {
		log(scAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     fmt.Sprintf("%v", err),
		})

		return
	}

	transactions = append(transactions, transaction)
	jsonString, err := json.Marshal(transactions)
	if err != nil {
		log(scAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     fmt.Sprintf("%v", err),
		})

		return
	}

	TxDB.Put(scAddress, string(jsonString))
}*/

//func newTxs(nonce int64, jsonString []byte) {
//	TxsDB.Put(strconv.FormatInt(nonce, 10), string(jsonString))
//}

/*func log(scAddress string, log contracts.Log) {
	scAddressLogs := LogDB.Get(scAddress).Value
	var logs []contracts.Log

	_ = json.Unmarshal([]byte(scAddressLogs), &logs)
	logs = append(logs, log)

	jsonString, _ := json.Marshal(logs)

	LogDB.Put(scAddress, string(jsonString))
}*/

/*func getScTokenBalance(scAddress string, tokenLabel string) float64 {
	if !crypt.IsAddressSmartContract(scAddress) || scAddress == "" || tokenLabel == "" {
		return 0
	}

	balance := contracts.GetAddress(scAddress).Balance
	if balance == nil {
		return 0
	}

	for _, i := range balance {
		if i.TokenLabel == tokenLabel {
			return i.Amount
		}
	}

	return 0
}*/

/*func refundTx(args DonateTokenConSmartArgs) {
	timestamp := apparel.TimestampUnix()
	timestampD := strconv.FormatInt(timestamp, 10)
	nonce := apparel.GetNonce(apparel.Timestamp())
	commentSign, err := json.Marshal(contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	))
	if err != nil {
		log(args.SenderAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     fmt.Sprintf("%v", err),
		})

		return
	}

	transaction := contracts.NewTx(
		5,
		nonce,
		"",
		config.BlockHeight,
		args.ScAddress,
		args.SenderAddress,
		args.Amount,
		config.BaseToken,
		timestampD,
		0,
		crypt.SignMessageWithSecretKey(
			config.NodeSecretKey,
			[]byte(config.NodeNdAddress),
		),
		*contracts.NewComment(
			"refund_transaction",
			commentSign,
		),
	)

	jsonString, err := json.Marshal(transaction)
	if err != nil {
		log(args.SenderAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     fmt.Sprintf("%v", err),
		})

		return
	}

	if memory.IsNodeProposer() {
		contracts.SendTx(jsonString)
		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
	}
}*/

/*func GetTxs(scAddress string) []contracts.Tx {
	var txs []contracts.Tx

	row := TxDB.Get(scAddress)
	err := json.Unmarshal([]byte(row.Value), &txs)
	if err != nil {
		return nil
	}

	if txs != nil {
		return sortForTimestamp(txs)
	} else {
		return nil
	}
}*/

/*func sortForTimestamp(txs []contracts.Tx) []contracts.Tx {
	sort.Slice(txs, func(i, j int) (less bool) {
		timestamp1, _ := strconv.ParseInt(txs[i].Timestamp, 10, 64)
		timestamp2, _ := strconv.ParseInt(txs[j].Timestamp, 10, 64)

		return timestamp1 < timestamp2
	})

	return txs
}*/

/*func ChangeStandard(scAddress string) error {
	scBalance := contracts.GetBalance(scAddress)
	if scBalance == nil {
		return nil
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return errors.New("Donate token contract error 1: token does not exist")
	}

	var txs []contracts.Tx
	for _, i := range scBalance {
		timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)

		tx := contracts.Tx{
			Type:       5,
			Nonce:      apparel.GetNonce(timestampD),
			HashTx:     "",
			Height:     config.BlockHeight,
			From:       scAddress,
			To:         token.Proposer,
			Amount:     i.Amount,
			TokenLabel: i.TokenLabel,
			Timestamp:  timestampD,
			Tax:        0,
			Signature:  crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
			Comment: *contracts.NewComment(
				"refund_transaction",
				nil,
			),
		}
		txs = append(txs, tx)
	}

	if txs != nil && memory.IsNodeProposer() {
		for _, i := range txs {
			transaction := contracts.NewTx(
				i.Type,
				i.Nonce,
				i.HashTx,
				i.Height,
				i.From,
				i.To,
				i.Amount,
				i.TokenLabel,
				i.Timestamp,
				i.Tax,
				i.Signature,
				i.Comment,
			)

			jsonString, _ := json.Marshal(transaction)

			contracts.SendTx(jsonString)
			*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
		}
	}

	return nil
}*/
