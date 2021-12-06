package blockchain

import (
	"encoding/json"
	"fmt"
	"log"
	"node/apparel"
	"node/blockchain/contracts/business_token_con"
	"node/blockchain/contracts/custom_turing_token_con"
	"node/blockchain/contracts/default_con"
	"node/blockchain/contracts/delegate_con"
	"node/blockchain/contracts/donate_token_con"
	"node/blockchain/contracts/holder_con"
	"node/blockchain/contracts/my_token_con"
	"node/blockchain/contracts/trade_token_con"
	"node/blockchain/contracts/vote_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"node/storage/validation"
	"node/websocket/sender"
	"strconv"
	"time"
)

type NodeOperation struct {
	PrevOperation int64
	Operation     int64
	Status        bool
}

var NodeOperationMemory NodeOperation

func Init() {
	log.Println("Block generator started.")

	Worker()
}

func Worker() {
	queue := make(chan int64)
	i := getQueue()
	go func() {
		for {
			if i != getQueue() {
				queue <- getQueue()
				i = getQueue()
			}
		}
	}()

	for {
		select {
		case j := <-queue:
			switch j {
			case config.BlockWriteTime[0]:
				if !memory.IsValidator() {
					if !memory.DownloadBlocks {
						sender.DownloadBlocksFromNodes()
					}
				} else {
					NodeOperationMemory.PrevOperation = 0
				}
				if NodeOperationMemory.PrevOperation == 0 {
					NodeOperationMemory.Operation = 1
					NodeOperationMemory.Status = false
					if memory.IsNodeProposer() && memory.IsValidator() {
						if !storage.CheckBlock(config.BlockHeight - 1) {
							config.BlockHeight--
							deep_actions.ConfigUpdate("block_height", strconv.FormatInt(config.BlockHeight,
								10))
						}

						log.Println("__PROPOSER__", memory.Proposer, "__PROPOSER__")
						log.Println("Block height:", config.BlockHeight)

						votes := addNodesForVote()

						var body deep_actions.Txs

						if int64(len(storage.TransactionsMemory)) <= config.MaxStorageMemory {
							for _, t := range storage.TransactionsMemory {
								body = append(body, t)
							}
						} else {
							for _, t := range storage.TransactionsMemory[:config.MaxStorageMemory] {
								body = append(body, t)
							}
						}

						rewardTransaction := rewardTransaction()
						if rewardTransaction.Amount != 0 && rewardTransaction.Amount >= 0 {
							body = append(body, rewardTransaction)
						}

						if config.BlockHeight%int64(len(memory.ValidatorsMemory)) == 0 {
							delegateTransaction := delegateTransaction()
							if delegateTransaction.Amount != 0 && delegateTransaction.Amount >= 0 {
								body = append(body, delegateTransaction)
							}
						}

						if votes != nil {
							block := storage.Block{
								Height:            config.BlockHeight,
								PrevHash:          deep_actions.GetChain(config.BlockHeight - 1).Hash,
								Timestamp:         strconv.FormatInt(apparel.TimestampUnix(), 10),
								Proposer:          config.NodeNdAddress,
								ProposerSignature: nil,
								Body:              body,
								Votes:             votes,
							}

							jsonString, _ := json.Marshal(block)
							block.ProposerSignature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)
							storage.BlockMemory = block

							sender.SendNewBlock()
						}
					}
					NodeOperationMemory.Status = true
					NodeOperationMemory.PrevOperation = 1
				}
				break
			case config.BlockWriteTime[1]:
				if NodeOperationMemory.PrevOperation == 1 && NodeOperationMemory.Status == true {
					NodeOperationMemory.Operation = 2
					NodeOperationMemory.Status = false
					if memory.IsValidator() {
						nodeVote := validateBlock(storage.BlockMemory)

						if storage.BlockMemory.Votes != nil {
							voteIdx := len(storage.BlockMemory.Votes)
							for idx, vote := range storage.BlockMemory.Votes {
								if vote.Proposer == config.NodeNdAddress {
									voteIdx = idx
									break
								}
							}

							if voteIdx != len(storage.BlockMemory.Votes) {

								vote := deep_actions.Vote{
									Proposer:    config.NodeNdAddress,
									Signature:   nil,
									BlockHeight: config.BlockHeight,
									Vote:        nodeVote,
								}

								jsonString, _ := json.Marshal(vote)

								vote.Signature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)

								storage.BlockMemory.Votes[voteIdx] = vote
								sender.SendBlockVote(vote)
							}
						}

						NodeOperationMemory.Status = true
						NodeOperationMemory.PrevOperation = 2
					}
				}
				break
			case config.BlockWriteTime[2]:
				if NodeOperationMemory.PrevOperation == 2 && NodeOperationMemory.Status == true {
					NodeOperationMemory.Operation = 3
					NodeOperationMemory.Status = false
					if memory.IsValidator() {
						if storage.BlockMemory.Votes != nil {

							if calculateVotes() {

								storage.AddBlock()

								stopArgs := vote_con.NewStopArgs(config.BlockHeight, storage.BlockMemory.Timestamp)

								err := stopArgs.Stop()
								if err != nil {
									log.Println("stop votes error: ", err)
								}

								for _, t := range storage.BlockMemory.Body {
									t.Amount = apparel.Round(t.Amount)
									t.Tax = apparel.Round(t.Tax)
									switch t.Type {
									case 1:
										ExecutionSmartContractsWithType1Transaction(t)
										break
									case 2:
										ExecutionSmartContractsWithType2Transaction(t)
										break
									case 3:
										ExecutionSmartContractsWithType3Transaction(t)
										break
									case 5:
										ExecutionSmartContractsWithType5Transaction(t)
										break
									}
								}
								log.Println("Block written")
							} else {
								log.Println("Block not written")
							}

							storage.BlockHeightUpdate()
						}

						NodeOperationMemory.Status = true
						NodeOperationMemory.PrevOperation = 3
					}
				}
				break
			case config.BlockWriteTime[3]:
				if NodeOperationMemory.PrevOperation == 3 && NodeOperationMemory.Status == true {
					NodeOperationMemory.Operation = 4
					NodeOperationMemory.Status = false
					if memory.IsValidator() {

						storage.ClearTransactionMemory()

						NodeOperationMemory.Status = true
						NodeOperationMemory.PrevOperation = 0
					}
				}

				memory.DownloadValidators()

				storage.BlockMemory = storage.Block{}
				memory.Proposer = memory.GetNextProposer()

				break
			}
		}
	}
}

func getQueue() int64 {
	return ((apparel.TimestampUnix() / time.Second.Nanoseconds()) % config.BlockWriteTime[3]) + 1
}

func calculateVotes() bool {
	var votes []deep_actions.Vote
	for _, vote := range storage.BlockMemory.Votes {
		if vote.Vote {
			votes = append(votes, vote)
		}
	}

	rule51 := calculate51(votes)
	rule66 := calculate66(votes)

	fmt.Printf("~ Voting is over ~\n")
	fmt.Printf("|%-6s|%-6t|\n", " 51% rule ", rule51)
	fmt.Printf("|%-6s|%-6t|\n", " 66% rule ", rule66)
	fmt.Printf("~\n")

	fmt.Printf("~ Voting nodes votes ~\n")

	fmt.Printf("|%-61s|%-20s|%-6s|\n", "Proposer", "Block Height", "Vote")
	for _, vote := range storage.BlockMemory.Votes {
		fmt.Printf("|%-61s|%-20d|%-6t|\n", vote.Proposer, vote.BlockHeight, vote.Vote)
	}

	fmt.Printf("~\n")
	return rule51 && rule66
}

func calculate51(votes []deep_actions.Vote) bool {
	return (len(votes) * 100 / len(storage.BlockMemory.Votes)) >= 51
}

func calculate66(votes []deep_actions.Vote) bool {
	var allNodesBalance = storage.GetAllNodesBalances()
	var trueNodesBalance float64
	for _, vote := range votes {
		for _, item := range storage.GetBalance(vote.Proposer) {
			if item.TokenLabel == config.RewardTokenLabel {
				trueNodesBalance += item.Amount
				break
			}
		}
	}

	return trueNodesBalance >= (allNodesBalance*66)/100
}

func rewardTransaction() deep_actions.Tx {
	amount := apparel.Round(storage.CalculateReward(config.NodeNdAddress))
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	comment := deep_actions.Comment{
		Title: "reward_transaction",
		Data:  nil,
	}

	tx := deep_actions.Tx{
		Type:       2,
		Nonce:      apparel.GetNonce(timestamp),
		Height:     config.BlockHeight,
		From:       config.GenesisAddress,
		To:         config.NodeUwAddress,
		Amount:     amount,
		TokenLabel: config.RewardTokenLabel,
		Timestamp:  timestamp,
		Comment:    comment,
	}

	tx.SetSignature(config.GenesisSecretKey)
	tx.SetHash()

	return tx
}

func delegateTransaction() deep_actions.Tx {
	amount := apparel.Round(storage.CalculateReward(config.DelegateScAddress))
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	comment := deep_actions.Comment{
		Title: "delegate_reward_transaction",
		Data:  nil,
	}

	tx := deep_actions.Tx{
		Type:       2,
		Nonce:      apparel.GetNonce(timestamp),
		Height:     config.BlockHeight,
		From:       config.GenesisAddress,
		To:         config.DelegateScAddress,
		Amount:     amount,
		TokenLabel: config.RewardTokenLabel,
		Timestamp:  timestamp,
		Comment:    comment,
	}

	tx.SetSignature(config.GenesisSecretKey)
	tx.SetHash()

	return tx
}

func addNodesForVote() deep_actions.Votes {
	var votes deep_actions.Votes

	if memory.ValidatorsMemory != nil {
		for _, node := range memory.ValidatorsMemory {
			votes = append(votes, deep_actions.Vote{Proposer: node.Address, BlockHeight: config.BlockHeight})
		}
	}

	return votes
}

func validateBlock(block storage.Block) bool {
	if err := validation.ValidateBlock(block); err != nil {
		log.Println(err)
		return false
	}

	return true
}

func ExecutionSmartContractsWithType1Transaction(t deep_actions.Tx) {
	commentData := make(map[string]interface{})
	_ = json.Unmarshal(t.Comment.Data, &commentData)

	switch t.Comment.Title {
	case "default_transaction":
		// pass
		break
	case "delegate_contract_transaction":
		delegateArgs, err := delegate_con.NewDelegateArgs(t.From, t.Amount)
		if err != nil {
			log.Println("Deep actions new tx delegate contract delegate transaction error 1:", err)
			break
		}

		err = delegate_con.Delegate(delegateArgs)
		if err != nil {
			log.Println("Deep actions new tx delegate contract delegate transaction error 2:", err)
			break
		}
		break
	case "undelegate_contract_transaction":
		delegateArgs, err := delegate_con.NewDelegateArgs(t.From, apparel.ConvertInterfaceToFloat64(commentData["undelegate_amount"]))
		if err != nil {
			log.Println("Deep actions new tx delegate contract undelegate transaction error 2:", err)
			break
		}

		err = delegate_con.SendUnDelegate(delegateArgs)
		if err != nil {
			log.Println("Deep actions new tx delegate contract undelegate transaction error 3:", err)
			break
		}
		break
	case "my_token_contract_confirmation_transaction":
		confirmationArgs, err := my_token_con.NewConfirmationArgs(t.To, t.From,
			t.Height, t.HashTx)
		if err != nil {
			log.Println(
				"Deep actions new tx confirmation transaction error 1:", err)
			break
		}

		err = my_token_con.Confirmation(confirmationArgs)
		if err != nil {
			log.Println("Deep actions new tx confirmation transaction error 2:", err)
		}

		break
	case "my_token_contract_get_percent_transaction":
		getPercentAgs, err := my_token_con.NewGetPercentArgs(t.To, t.From,
			t.TokenLabel, t.Amount, t.Height, t.HashTx)
		if err != nil {
			log.Println("Deep actions new tx my contract get percent transaction error 1:", err)
			break
		}

		err = my_token_con.GetPercent(getPercentAgs)
		if err != nil {
			log.Println("Deep actions new tx my contract get percent transaction error 2:", err)
		}

		break
	case "donate_token_contract_buy_transaction":
		buyAgs, err := donate_token_con.NewBuyArgs(t.To, t.From,
			t.Amount, t.HashTx, t.Height)
		if err != nil {
			log.Println("Deep actions new tx donate contract buy transaction error 1:", err)
			break
		}

		err = donate_token_con.Buy(buyAgs)
		if err != nil {
			log.Println("Deep actions new tx donate contract buy transaction error 2:", err)
		}

		break
	case "donate_token_contract_fill_config_transaction":
		scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, t.From)
		fillConfigAgs, err := donate_token_con.NewFillConfigArgs(scAddress,
			apparel.ConvertInterfaceToFloat64(commentData["conversion"]),
			apparel.ConvertInterfaceToFloat64(commentData["max_buy"]), apparel.ConvertInterfaceToBool(commentData["changes"]), t.HashTx, t.Height)
		if err != nil {
			log.Println("Deep actions new tx donate contract fill config transaction error 1:", err)
			break
		}

		if err = fillConfigAgs.FillConfig(); err != nil {
			log.Println("Deep actions new tx donate contract fill config transaction error 2:", err)
		}

		break
	case "business_token_contract_buy_transaction":
		buyArgs, err := business_token_con.NewBuyArgs(t.To, t.From, t.TokenLabel,
			t.Amount, t.HashTx, t.Height)
		if err != nil {
			log.Println(
				"Deep actions new tx business contract buy transaction error 1:",
				err)
			break
		}

		err = business_token_con.Buy(buyArgs)
		if err != nil {
			log.Println(
				"Deep actions new tx business contract buy transaction error 2:",
				err)
		}
		break
	case "business_token_contract_get_percent_transaction":
		buyArgs, err := business_token_con.NewGetPercentArgs(t.To, t.From, apparel.ConvertInterfaceToFloat64(commentData["amount"]),
			apparel.ConvertInterfaceToString(commentData["token_label"]), t.HashTx, t.Height)
		if err != nil {
			log.Println(
				"Deep actions new tx business contract get percent transaction error 1:",
				err)
			break
		}

		err = buyArgs.GetPercent()
		if err != nil {
			log.Println(
				"Deep actions new tx business contract get percent transaction error 2:",
				err)
		}
		break
	case "business_token_contract_fill_config_transaction":
		scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, t.From)
		partners := business_token_con.Partners{}
		jsonString, _ := json.Marshal(commentData["partners"])
		_ = json.Unmarshal(jsonString, &partners)
		fillConfigArgs, err := business_token_con.NewFillConfigArgs(scAddress,
			apparel.ConvertInterfaceToFloat64(commentData["conversion"]),
			apparel.ConvertInterfaceToFloat64(commentData["sales_value"]),
			partners, apparel.ConvertInterfaceToBool(commentData["changes"]), t.HashTx, t.Height)
		if err != nil {
			log.Println(
				"Deep actions new tx business contract fill config transaction error 1:",
				err)
			break
		}

		err = fillConfigArgs.FillConfig()
		if err != nil {
			log.Println(
				"Deep actions new tx business contract fill config transaction error 2:",
				err)
		}
		break
	case "trade_token_contract_add_transaction":
		err := trade_token_con.Add(trade_token_con.NewTradeArgs(t.To, t.From,
			t.Amount, t.TokenLabel, t.Height, t.HashTx))
		if err != nil {
			log.Println(
				"Deep actions new tx trade contract add transaction error 1:",
				err)
			break
		}
		break
	case "trade_token_contract_swap_transaction":
		err := trade_token_con.Swap(trade_token_con.NewTradeArgs(t.To, t.From,
			t.Amount, t.TokenLabel, t.Height, t.HashTx))
		if err != nil {
			log.Println(
				"Deep actions new tx trade contract swap transaction error 1:",
				err)
			break
		}
		break
	case "trade_token_contract_get_liq_transaction":
		err := trade_token_con.GetLiq(trade_token_con.NewGetArgs(t.To, t.From,
			t.TokenLabel, t.Height, t.HashTx))
		if err != nil {
			log.Println(
				"Deep actions new tx trade contract get transaction error 1:",
				err)
			break
		}
		break
	case "trade_token_contract_get_com_transaction":
		err := trade_token_con.GetCom(trade_token_con.NewGetArgs(t.To, t.From,
			t.TokenLabel, t.Height, t.HashTx))
		if err != nil {
			log.Println(
				"Deep actions new tx trade contract get transaction error 1:",
				err)
			break
		}
		break
	case "trade_token_contract_fill_config_transaction":
		scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, t.From)
		fillConfigArgs, err := trade_token_con.NewFillConfigArgs(scAddress,
			apparel.ConvertInterfaceToFloat64(commentData["commission"]),
			apparel.ConvertInterfaceToBool(commentData["changes"]),
			t.HashTx, t.Height)
		if err != nil {
			log.Println("Deep actions new tx trade token contract fill config transaction error 1:", err)
			break
		}

		if err := fillConfigArgs.FillConfig(); err != nil {
			log.Println(
				"Deep actions new tx trade token contract fill config transaction error 2:",
				err)
			break
		}
		break
	case "holder_contract_add_transaction":
		holderAddArgs, err := holder_con.NewAddArgs(t.From,
			apparel.ConvertInterfaceToString(commentData["get_tokens_address"]),
			t.Amount,
			t.TokenLabel,
			apparel.ConvertInterfaceToInt64(commentData["get_block_height"]),
			t.HashTx, t.Height)
		if err != nil {
			log.Println("Deep actions new tx holder contract add transaction error 2:", err)
			break
		}

		err = holderAddArgs.Add()
		if err != nil {
			log.Println("Deep actions new tx holder contract add transaction error 3:", err)
			break
		}

		break
	case "holder_contract_get_transaction":
		holderGetArgs, err := holder_con.NewGetArgs(t.From, t.HashTx, t.Height)
		if err != nil {
			log.Println(
				"Deep actions new tx holder contract get transaction error 1:",
				err)
			break
		}

		err = holder_con.Get(holderGetArgs)
		if err != nil {
			log.Println(
				"Deep actions new tx holder contract get transaction error 2:",
				err)
			break
		}

		break
	case "vote_contract_start_transaction":
		voteStartArgs, err := vote_con.NewStartArgs(
			apparel.ConvertInterfaceToString(commentData["title"]),
			apparel.ConvertInterfaceToString(commentData["description"]),
			commentData["answer_options"],
			apparel.ConvertInterfaceToInt64(commentData["end_block_height"]),
			t.From, t.HashTx, t.Height)
		if err != nil {
			log.Println(
				"Deep actions new tx vote contract start transaction error 2:",
				err)
			break
		}

		err = vote_con.Start(voteStartArgs)
		if err != nil {
			log.Println(
				"Deep actions new tx vote contract start transaction error 3:",
				err)
			break
		}

		break
	case "vote_contract_hard_stop_transaction":
		voteHardStopArgs, err := vote_con.NewHardStopArgs(apparel.ConvertInterfaceToInt64(commentData["vote_nonce"]), t.HashTx, t.Height, t.From)
		if err != nil {
			log.Println(
				"Deep actions new tx vote contract hard stop transaction error 2:",
				err)
			break
		}

		err = vote_con.HardStop(voteHardStopArgs)
		if err != nil {
			log.Println(
				"Deep actions new tx vote contract hard stop transaction error 3:",
				err)
			break
		}

		break
	case "vote_contract_answer_transaction":
		voteAnswerArgs, err := vote_con.NewAnswerArgs(t.From, t.Signature, t.HashTx,
			apparel.ConvertInterfaceToString(commentData["possible_answer_nonce"]),
			apparel.ConvertInterfaceToInt64(commentData["vote_nonce"]), t.Height)
		if err != nil {
			log.Println(
				"Deep actions new tx vote contract answer transaction error 2:",
				err)
			break
		}

		err = voteAnswerArgs.Answer()
		if err != nil {
			log.Println(
				"Deep actions new tx vote contract answer transaction error 3:",
				err)
			break
		}

		break
	case "custom_turing_token_add_emission_transaction":
		if err := custom_turing_token_con.AddEmission(
			custom_turing_token_con.NewAddEmissionArgs(
				apparel.ConvertInterfaceToFloat64(commentData["add_emission_amount"]),
				t.Height, t.HashTx)); err != nil {
			log.Println(
				"Deep actions new tx custom turing token contract add emission transaction error 1:",
				err)
			break
		}

		break
	case "custom_turing_token_de_delegate_transaction":
		if err := custom_turing_token_con.DeDelegate(
			custom_turing_token_con.NewDeDelegateArgs(t.From,
				apparel.ConvertInterfaceToFloat64(commentData["de_delegate_amount"]),
				t.Height, t.HashTx)); err != nil {
			log.Println(
				"Deep actions new tx custom turing token contract de-delegate transaction error 1:",
				err)
			break
		}

		break
	case "custom_turing_token_de_delegate_another_address_transaction":
		if err := custom_turing_token_con.DeDelegateAnotherAddress(
			custom_turing_token_con.NewDeDelegateAnotherAddressArgs(t.From,
				apparel.ConvertInterfaceToString(commentData["de_delegate_recipient_address"]),
				apparel.ConvertInterfaceToFloat64(commentData["de_delegate_amount"]),
				t.Height, t.HashTx)); err != nil {
			log.Println(
				"Deep actions new tx custom turing token contract de-delegate to another address transaction error 1:",
				err)
			break
		}

		break
	case "custom_turing_token_get_reward_transaction":
		if err := custom_turing_token_con.GetReward(
			custom_turing_token_con.NewGetRewardArgs(t.Height, t.HashTx)); err != nil {
			log.Println("Deep actions new tx custom turing token contract get reward transaction error 1:", err)
			break
		}

		break
	case "custom_turing_token_re_delegate_transaction":
		if err := custom_turing_token_con.ReDelegate(
			custom_turing_token_con.NewReDelegateArgs(t.From,
				apparel.ConvertInterfaceToString(commentData["re_delegate_recipient_address"]),
				apparel.ConvertInterfaceToFloat64(commentData["re_delegate_amount"]),
				t.Height, t.HashTx)); err != nil {
			log.Println("Deep actions new tx custom turing token contract re-delegate transaction error 1:", err)
			break
		}

		break
	case "custom_turing_token_delegate_transaction":
		if err := custom_turing_token_con.Delegate(
			custom_turing_token_con.NewDelegateArgs(t.From, t.Amount, t.Height, t.HashTx)); err != nil {
			log.Println("Deep actions new tx custom turing token contract delegate transaction error 1:", err)
			break
		}

		break
	case "default_contract_create_transaction":
		var defaultContractCreateCommentData []interface{}
		err := json.Unmarshal(t.Comment.Data, &defaultContractCreateCommentData)
		if err != nil {
			log.Println("Deep actions new tx default contract create transaction error 1:", err)
			break
		}

		if defaultContractCreateCommentData == nil {
			log.Println("Deep actions new tx default contract create transaction error 2:", err)
			break
		}

		for i := range defaultContractCreateCommentData {
			el := apparel.ConvertInterfaceToMapStringInterface(defaultContractCreateCommentData[i])
			createArgs, err := default_con.NewCreate(t.From, apparel.ConvertInterfaceToString(el["name"]),
				apparel.ConvertInterfaceToFloat64(el["price"]),
				apparel.ConvertInterfaceToString(el["data"]), t.HashTx, t.Height)
			if err != nil {
				log.Println("Deep actions new tx default contract create transaction error 3:", err)
				break
			}

			if err := createArgs.Create(); err != nil {
				log.Println("Deep actions new tx default contract create transaction error 4:", err)
				break
			}
		}

		break
	case "default_contract_buy_transaction":
		createArgs, err := default_con.NewBuy(apparel.ConvertInterfaceToInt64(commentData["id"]), t.HashTx,
			t.From, t.Amount, t.Height)
		if err != nil {
			log.Println("Deep actions new tx default contract buy transaction error 2:", err)
			break
		}

		if err := createArgs.Buy(); err != nil {
			log.Println("Deep actions new tx default contract buy transaction error 3:", err)
			break
		}

		break
	case "default_contract_set_price_transaction":
		setPriceArgs, err := default_con.NewSetPrice(apparel.ConvertInterfaceToInt64(commentData["id"]),
			apparel.ConvertInterfaceToFloat64(commentData["new_price"]), t.HashTx, t.Height)
		if err != nil {
			log.Println("Deep actions new tx default contract set price transaction error 2:", err)
			break
		}

		if err := setPriceArgs.SetPrice(); err != nil {
			log.Println("Deep actions new tx default contract set price transaction error 3:", err)
			break
		}
		break
	case "default_contract_fill_config_transaction":
		scAddress := crypt.AddressFromAnotherAddress(metrics.SmartContractPrefix, t.From)
		fillConfigArgs, err := default_con.NewFillConfigArgs(scAddress,
			apparel.ConvertInterfaceToFloat64(commentData["commission"]), t.HashTx, t.Height)
		if err != nil {
			log.Println("Deep actions new tx trade token contract fill config transaction error 1:", err)
			break
		}

		if err := fillConfigArgs.FillConfig(); err != nil {
			log.Println("Deep actions new tx trade token contract fill config transaction error 2:", err)
			break
		}
		break
	}
}

func ExecutionSmartContractsWithType2Transaction(t deep_actions.Tx) {
	if t.Comment.Title == "delegate_reward_transaction" && t.To == config.DelegateScAddress {
		_ = delegate_con.Bonus(t.Timestamp)
	}
}

func ExecutionSmartContractsWithType3Transaction(t deep_actions.Tx) {
	address := deep_actions.GetAddress(t.From)
	token := address.GetToken()
	switch t.Comment.Title {
	case "change_token_standard_transaction":
		if token.Id == 0 {
			break
		}

		var standard int64 = 0
		if token.StandardHistory != nil {
			if len(token.StandardHistory) != 1 {
				hash := token.StandardHistory[len(token.StandardHistory)-1].TxHash
				if hash == "" {
					log.Println("Deep actions new tx change token standard transaction error 1")
					break
				}

				jsonString := storage.GetTxForHash(hash)
				if jsonString == "" {
					log.Println("Deep actions new tx change token standard transaction error 2")
					break
				}

				tx := deep_actions.Tx{}
				err := json.Unmarshal([]byte(jsonString), &tx)
				if err != nil {
					log.Println("Deep actions new tx change token standard transaction error 3:", err)
					break
				}

				if tx.Comment.Title != "change_token_standard_transaction" {
					log.Println("Deep actions new tx change token standard transaction error 4")
					break
				}

				t := deep_actions.Token{}
				err = json.Unmarshal(tx.Comment.Data, &t)
				if err != nil {
					log.Println("Deep actions new tx change token standard transaction error 5:",
						err)
					break
				}

				standard = t.Standard

				if standard == token.Standard {
					hash := token.StandardHistory[len(token.StandardHistory)-2].TxHash
					if hash == "" {
						log.Println("Deep actions new tx change token standard transaction error 1")
						break
					}

					jsonString := storage.GetTxForHash(hash)
					if jsonString == "" {
						log.Println("Deep actions new tx change token standard transaction error 2")
						break
					}

					tx := deep_actions.Tx{}
					err := json.Unmarshal([]byte(jsonString), &tx)
					if err != nil {
						log.Println("Deep actions new tx change token standard transaction error 3:", err)
						break
					}

					if tx.Comment.Title != "change_token_standard_transaction" {
						log.Println("Deep actions new tx change token standard transaction error 4")
						break
					}

					t := deep_actions.Token{}
					err = json.Unmarshal(tx.Comment.Data, &t)
					if err != nil {
						log.Println("Deep actions new tx change token standard transaction error 5:", err)
						break
					}

					standard = t.Standard
				}
			}
		}

		publicKey, err := crypt.PublicKeyFromAddress(t.From)
		if err != nil {
			log.Println("Deep actions new tx change token standard transaction error 6:", err)
			break
		}

		scAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)
		switch standard {
		case 0:
			err := my_token_con.ChangeStandard(scAddress)
			if err != nil {
				log.Println("Deep actions new tx change token standard transaction error 7:", err)
				break
			}
			break
		case 1:
			err := donate_token_con.ChangeStandard(scAddress)
			if err != nil {
				log.Println("Deep actions new tx change token standard transaction error 8:", err)
				break
			}
			break
		case 4:
			err := business_token_con.ChangeStandard(scAddress)
			if err != nil {
				log.Println("Deep actions new tx change token standard transaction error 9:", err)
				break
			}
			break
		}

		if token.Standard == 5 {
			err := trade_token_con.AddToken(scAddress)
			if err != nil {
				log.Println("Deep actions new tx change token standard transaction error 10:", err)
			}
		}
		break
	}
}

func ExecutionSmartContractsWithType5Transaction(t deep_actions.Tx) {
	switch t.Comment.Title {
	case "undelegate_contract_transaction":
		if t.From != config.DelegateScAddress {
			log.Println("Deep actions new tx delegate contract undelegate transaction error 1")
			break
		}

		commentData := make(map[string]interface{})
		err := json.Unmarshal(t.Comment.Data, &commentData)
		if err != nil {
			log.Println("Deep actions new tx delegate contract undelegate transaction error 2:", err)
			break
		}

		delegateArgs, err := delegate_con.NewDelegateArgs(t.To, t.Amount)
		if err != nil {
			log.Println("Deep actions new tx delegate contract undelegate transaction error 3:", err)
			break
		}

		err = delegate_con.UnDelegate(delegateArgs)
		if err != nil {
			log.Println("Deep actions new tx delegate contract undelegate transaction error 4:", err)
		}

		break
	}
}
