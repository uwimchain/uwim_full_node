package my_token_con

import (
	"node/blockchain/contracts"
)

var (
	db = contracts.Database{}

	PoolDB = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_pool")
	//TxDB    = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_tx")
	//TxsDB   = db.NewConnection("blockchain/contracts/donate_token_con/storage/donate_token_contract_txs")
	//LogDB   = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_log")
	EventDB  = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/my_token_con/storage/my_token_contract_config")
)

type Pool struct {
	Address string `json:"address"`
}

/*type MyTokenConSmartArgs struct {
	ScAddress     string  `json:"sc_address"`
	SenderAddress string  `json:"sender_address"`
	Amount        float64 `json:"amount"`
	TokenLabel    string  `json:"token_label"`
}*/

/*func Smart(args MyTokenConSmartArgs) error {
	timestamp := apparel.TimestampUnix()
	timestampD := apparel.UnixToString(timestamp)

	if !crypt.IsAddressSmartContract(args.ScAddress) || args.ScAddress == "" {
		log(args.ScAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "sc address is null or not sc address",
		})

		return errors.New("sc address is null or not sc address")
	}

	scAddress := args.ScAddress

	if !crypt.IsAddressUw(args.SenderAddress) || args.SenderAddress == "" {
		log(scAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "sender address is null or not uwim address",
		})

		return errors.New("sender address is null or not uwim address")
	}

	if args.Amount <= 0 {
		log(scAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "null or negative amount",
		})

		return errors.New("null or negative amount")
	}

	if args.TokenLabel == config.BaseToken {
		return nil
	}

	scToken := contracts.GetTokenInfoForScAddress(scAddress)
	if scToken.Id == 0 {
		log(scAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     "this token does not exist",
		})

		return errors.New("this token does not exist")
	}

	pool := getPoolForSc(scAddress)
	if pool == nil {
		//	log(scAddress, contracts.Log{
		//		Timestamp:  timestamp,
		//		TimestampD: timestampD,
		//		Record:     "pool of this scToken is null",
		//	})
		//
		//	return errors.New("pool of this scToken is null")
		return nil
	}

	if args.TokenLabel == scToken.Label {
		//addressPercent := getPoolAddressPercent(scAddress, pool, args.SenderAddress, scToken.Emission, args.Amount, scToken.Label)
		addressPercent := getPoolAddressPercent(GetPoolAddressPercentArgs{
			ScAddress: scAddress,
			UwAddress: args.SenderAddress,
			Amount:    args.Amount,
		})

		if addressPercent == 0 {

			refundTx(args)
			return errors.New("address percent is null")
		}

		scBalance := contracts.GetAddress(scAddress).Balance
		if scBalance == nil {

			refundTx(args)
			return errors.New("address sc balance is null")
		}

		var transactions []contracts.Tx
		for _, i := range scBalance {
			if i.TokenLabel != scToken.Label {
				amount := i.Amount * (addressPercent / 100)
				tax := apparel.CalcTax(amount * config.Tax)
				nonce := apparel.GetNonce(apparel.Timestamp())

				commentSign, err := json.Marshal(contracts.NewBuyTokenSign(
					config.NodeNdAddress,
				))
				if err != nil {
					log(scAddress, contracts.Log{
						Timestamp:  timestamp,
						TimestampD: timestampD,
						Record:     fmt.Sprintf("%v", err),
					})

					return err
				}

				transaction := contracts.Tx{
					Type:       5,
					Nonce:      nonce,
					HashTx:     "",
					Height:     config.BlockHeight,
					From:       scAddress,
					To:         args.SenderAddress,
					Amount:     amount,
					TokenLabel: i.TokenLabel,
					Timestamp:  apparel.UnixToString(timestamp),
					Tax:        tax,
					Signature: crypt.SignMessageWithSecretKey(
						config.NodeSecretKey,
						[]byte(config.NodeNdAddress),
					),
					Comment: *contracts.NewComment(
						"default_transaction",
						commentSign,
					),
				}

				transactions = append(transactions, transaction)
			}
		}

		if transactions == nil {
			return nil
		}

		addEvent(scAddress, Event{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Text:       fmt.Sprintf("%s take his percent %g", args.SenderAddress, addressPercent),
		})

		var allTax float64 = 0
		for _, t := range transactions {
			allTax += t.Tax

			for _, i := range scBalance {
				if i.TokenLabel == t.TokenLabel {
					if i.Amount <= t.Amount {
						log(scAddress, contracts.Log{
							Timestamp:  timestamp,
							TimestampD: timestampD,
							Record:     fmt.Sprintf("sc address low balance of %s for transaction", t.TokenLabel),
						})

						// refund tokens, if sc address has low balance for get a address percent
						refundTx(args)

						return errors.New("low balance on sc address")
					}

					break
				}
			}
		}

		for _, i := range scBalance {
			if i.TokenLabel == config.BaseToken {
				if i.Amount <= allTax {
					log(scAddress, contracts.Log{
						Timestamp:  timestamp,
						TimestampD: timestampD,
						Record:     "sc address low balance of uwim for get address percent transactions taxes",
					})

					// refund tokens, if sc address has low balance for get a address percent
					refundTx(args)

					return errors.New("sc address low balance of uwim for get address percent transactions taxes")
				}

				break
			}
		}

		if memory.IsNodeProposer() {
			for _, t := range transactions {
				jsonString, _ := json.Marshal(t)
				contracts.SendTx(jsonString)
				*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *contracts.NewTx(
					t.Type,
					t.Nonce,
					t.HashTx,
					t.Height,
					t.From,
					t.To,
					t.Amount,
					t.TokenLabel,
					t.Timestamp,
					t.Tax,
					t.Signature,
					t.Comment,
				))
			}
		}

		for _, t := range transactions {
			jsonString, _ := json.Marshal(t)
			newTx(scAddress, jsonString)
			newTxs(t.Nonce, jsonString)
		}
	}

	return nil
}*/

/*func Confirmation(scAddress string, address string) error {
	timestamp := apparel.TimestampUnix()
	timestampD := apparel.UnixToString(timestamp)
	pool := getPoolForSc(scAddress)

	if pool != nil {
		for _, i := range pool {
			if i.Address == address {
				return errors.New("this address already exists of this token pool")
			}
		}
	}

	pool = append(pool, Pool{
		Address: address,
	})

	jsonString, err := json.Marshal(pool)
	if err != nil {
		log(scAddress, contracts.Log{
			Timestamp:  timestamp,
			TimestampD: timestampD,
			Record:     fmt.Sprintf("%v", err),
		})

		return err
	}

	addEvent(scAddress, Event{
		Timestamp:  timestamp,
		TimestampD: timestampD,
		Text:       fmt.Sprintf("%s confirmation", address),
	})

	PoolDB.Put(scAddress, string(jsonString))

	return nil
}*/

/*func getPoolForSc(scAddress string) []Pool {
	if !crypt.IsAddressSmartContract(scAddress) {
		return nil
	}

	var pool []Pool
	jsonString := PoolDB.Get(scAddress).Value
	err := json.Unmarshal([]byte(jsonString), &pool)
	if err != nil {
		//timestamp := apparel.TimestampUnix()
		//timestampD := apparel.UnixToString(timestamp)
		//log(scAddress, contracts.Log{
		//	Timestamp:  timestamp,
		//	TimestampD: timestampD,
		//	Record:     fmt.Sprintf("%v", err),
		//})

		return pool
	}

	return pool
}*/

/*type GetPoolAddressPercentArgs struct {
	ScAddress string  `json:"sc_address"`
	UwAddress string  `json:"uw_address"`
	Amount    float64 `json:"amount"`
}*/

/*func getPoolAddressPercent(args GetPoolAddressPercentArgs) float64 {
	if !crypt.IsAddressSmartContract(args.ScAddress) {
		return 0
	}

	pool := getPoolForSc(args.ScAddress)
	token := contracts.GetTokenInfoForScAddress(args.ScAddress)
	if pool == nil {
		return 0
	}

	if !crypt.IsAddressUw(args.UwAddress) || args.UwAddress == "" {
		return 0
	}

	for _, i := range pool {
		if i.Address == args.UwAddress {
			scBalance := getScTokenBalance(args.ScAddress, token.Label)

			percent := args.Amount / (token.Emission - scBalance)
			return percent
		}
	}

	return 0
}*/

/*func GetAddressPercent(scAddress string, uwAddress string, tokenLabel string, emission float64, amount float64) (bool, float64, error) {
	pool := getPoolForSc(scAddress)

	if pool == nil {
		//return false, 0, errors.New("Invalid insert data pool is null")
		return false, 0, nil
	}

	if !crypt.IsAddressSmartContract(scAddress) {
		return false, 0, errors.New("Invalid insert data " + scAddress)
	}

	if !crypt.IsAddressUw(uwAddress) {
		return false, 0, errors.New("Invalid insert data " + uwAddress)
	}

	for _, i := range pool {
		if i.Address == uwAddress {
			scBalance := getScTokenBalance(scAddress, tokenLabel)
			percent := amount / (emission - scBalance)
			return true, percent, nil
		}
	}

	return false, 0, nil
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

/*func log(scAddress string, log contracts.Log) {
	scAddressLogs := LogDB.Get(scAddress).Value
	var logs []contracts.Log

	_ = json.Unmarshal([]byte(scAddressLogs), &logs)
	logs = append(logs, log)

	jsonString, _ := json.Marshal(logs)

	LogDB.Put(scAddress, string(jsonString))
}*/

/*type Event struct {
	Timestamp  int64  `json:"timestamp"`
	TimestampD string `json:"timestamp_d"`
	Text       string `json:"text"`
}

func addEvent(scAddress string, event Event) {
	var events []Event
	jsonEvents := EventDB.Get(scAddress).Value
	_ = json.Unmarshal([]byte(jsonEvents), &events)

	events = append(events, event)

	jsonString, _ := json.Marshal(events)
	EventDB.Put(scAddress, string(jsonString))
}*/

/*func refundTx(args MyTokenConSmartArgs) {
	if memory.IsNodeProposer() {
		timestamp := apparel.TimestampUnix()
		timestampD := strconv.FormatInt(timestamp, 10)
		scAddress := args.ScAddress
		nonce := apparel.GetNonce(apparel.Timestamp())
		commentSign, err := json.Marshal(contracts.NewBuyTokenSign(
			config.NodeNdAddress,
		))
		if err != nil {
			log(scAddress, contracts.Log{
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
			scAddress,
			args.SenderAddress,
			args.Amount,
			args.TokenLabel,
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
			log(scAddress, contracts.Log{
				Timestamp:  timestamp,
				TimestampD: timestampD,
				Record:     fmt.Sprintf("%v", err),
			})

			return
		}

		contracts.SendTx(jsonString)
		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
	}
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

/*func newTxs(nonce int64, jsonString []byte) {
	TxsDB.Put(strconv.FormatInt(nonce, 10), string(jsonString))
}*/

/*func ValidateConfirmation(scAddress string, uwAddress string) int64 {
	publicKey, err := crypt.PublicKeyFromAddress(uwAddress)
	if err != nil {
		return 101
	}

	if scAddress == crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey) {
		return 102
	}

	pool := getPoolForSc(scAddress)

	for _, i := range pool {
		if i.Address == uwAddress {
			return 103
		}
	}

	return 0
}*/

/*type PoolInfo struct {
	Address    string  `json:"address"`
	Amount     float64 `json:"amount"`
	TokenLabel string  `json:"token_label"`
	Percent    float64 `json:"percent"`
}

func GetPool(scAddress string) []PoolInfo {
	pool := getPoolForSc(scAddress)
	token := contracts.GetTokenInfoForScAddress(scAddress)
	var result []PoolInfo

	for _, i := range pool {
		addressBalance := contracts.GetBalanceForToken(i.Address, token.Label)

		result = append(result, PoolInfo{
			Address:    i.Address,
			Amount:     addressBalance.Amount,
			TokenLabel: token.Label,
			Percent: getPoolAddressPercent(GetPoolAddressPercentArgs{
				ScAddress: scAddress,
				UwAddress: i.Address,
				Amount:    addressBalance.Amount,
			}),
		})
	}

	return result
}*/

/*func ChangeStandard(scAddress string) {
	scBalance := contracts.GetBalance(scAddress)
	if scBalance == nil {
		return
	}

	pool := getPoolForSc(scAddress)
	//if pool == nil {
	//	return
	//}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return
	}

	var txs []contracts.Tx
	var allTxsAmount []contracts.Balance
	if pool != nil {
		for _, i := range pool {
			addressBalance := contracts.GetBalanceForToken(i.Address, token.Label)
			percent := getPoolAddressPercent(GetPoolAddressPercentArgs{
				ScAddress: scAddress,
				UwAddress: i.Address,
				Amount:    addressBalance.Amount,
			})

			if percent == 0 {
				continue
			}

			for _, j := range scBalance {
				if j.TokenLabel == token.Label {
					continue
				}

				timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)

				txs = append(txs, contracts.Tx{
					Type:       5,
					Nonce:      apparel.GetNonce(timestampD),
					HashTx:     "",
					Height:     config.BlockHeight,
					From:       scAddress,
					To:         i.Address,
					Amount:     j.Amount * (percent / 100),
					TokenLabel: j.TokenLabel,
					Timestamp:  timestampD,
					Tax:        0,
					Signature:  crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
					Comment:    *contracts.NewComment("refund_transaction", nil),
				})

				if allTxsAmount != nil {
					check := false
					for _, k := range allTxsAmount {
						if k.TokenLabel == j.TokenLabel {
							k.Amount += j.Amount
							check = true
							break
						}
					}

					if !check {
						allTxsAmount = append(allTxsAmount, contracts.Balance{
							TokenLabel: j.TokenLabel,
							Amount:     j.Amount,
						})
					}
				} else {
					allTxsAmount = append(allTxsAmount, contracts.Balance{
						TokenLabel: j.TokenLabel,
						Amount:     j.Amount,
					})
				}
			}
		}
	}

	if allTxsAmount != nil {
		for _, i := range allTxsAmount {
			for _, j := range scBalance {
				if i.TokenLabel == j.TokenLabel {
					timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)
					tx := contracts.Tx{
						Type:       5,
						Nonce:      apparel.GetNonce(timestampD),
						HashTx:     "",
						Height:     config.BlockHeight,
						From:       scAddress,
						To:         token.Proposer,
						Amount:     j.Amount - i.Amount,
						TokenLabel: j.TokenLabel,
						Timestamp:  timestampD,
						Tax:        0,
						Signature:  crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
						Comment: *contracts.NewComment(
							"refund_transaction",
							nil,
						),
					}
					txs = append(txs, tx)
					break
				}
			}
		}
	} else {
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
	}

	if txs != nil && memory.IsNodeProposer() {
		for _, j := range txs {

			transaction := contracts.NewTx(
				j.Type,
				j.Nonce,
				"",
				j.Height,
				j.From,
				j.To,
				j.Amount,
				j.TokenLabel,
				j.Timestamp,
				j.Tax,
				j.Signature,
				j.Comment,
			)

			jsonString, _ := json.Marshal(transaction)

			contracts.SendTx(jsonString)
			*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *transaction)
		}
	}
}*/
