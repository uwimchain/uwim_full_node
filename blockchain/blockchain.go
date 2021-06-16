package blockchain

import (
	"encoding/json"
	"fmt"
	"log"
	"node/apparel"
	"node/blockchain/contracts/delegate_con"
	"node/config"
	"node/crypt"
	"node/memory"
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
	sender.SendVersion(nil)
	sender.DownloadBlocksFromNodes()

	for {
		Worker()
	}
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
					sender.SendVersion(nil)

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
							storage.ConfigUpdate("block_height", strconv.FormatInt(config.BlockHeight, 10))
						}

						log.Println("__PROPOSER__", memory.Proposer, "__PROPOSER__")
						log.Println("Block height:", config.BlockHeight)

						votes := addNodesForVote()

						var body []deep_actions.Tx

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

						if config.BlockHeight%config.DelegateBlockHeight == 0 {
							miningTransaction := delegateTransaction()
							if miningTransaction.Amount != 0 && miningTransaction.Amount >= 0 {
								body = append(body, miningTransaction)
							}
						}

						if votes != nil {
							storage.BlockMemory = *storage.NewBlock(
								config.BlockHeight,
								storage.GetBlockHash(config.BlockHeight-1),
								apparel.Timestamp(),
								config.NodeNdAddress,
								crypt.SignMessageWithSecretKey(
									config.NodeSecretKey,
									[]byte(config.NodeNdAddress),
								),
								body,
								votes,
							)

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

								sign := crypt.SignMessageWithSecretKey(
									config.NodeSecretKey,
									[]byte(config.NodeNdAddress),
								)
								jsonString, err := json.Marshal(*deep_actions.NewVote(
									config.NodeNdAddress,
									sign,
									config.BlockHeight,
									nodeVote,
								))
								if err != nil {
									log.Println("Vote Sign error:", err)
								}

								storage.BlockMemory.Votes[voteIdx].Vote = nodeVote
								storage.BlockMemory.Votes[voteIdx].Signature = sign
								storage.BlockMemory.Votes[voteIdx].BlockHeight = config.BlockHeight

								sender.SendBlockVote(jsonString)
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

								for _, t := range storage.BlockMemory.Body {
									switch t.Type {
									case 1:
										{
											switch t.Comment.Title {
											case "delegate_contract_transaction":
												{
													if t.To == config.DelegateScAddress {
														timestampUnix := apparel.UnixFromStringTimestamp(t.Timestamp)
														err := delegate_con.Delegate(t.From, t.Amount, timestampUnix)
														if err != nil {
															log.Println("Deep actions new tx delegate contract transaction error 1:", err)
														}
													} else {
														log.Println("Deep actions new tx delegate contract transaction error 2")
													}
													break
												}
											case "undelegate_contract_transaction":
												{
													if t.To == config.DelegateScAddress {
														undelegateCommentData := delegate_con.UndelegateCommentData{}
														_ = json.Unmarshal(t.Comment.Data, &undelegateCommentData)
														err := delegate_con.SendUnDelegate(t.From, undelegateCommentData.Amount)
														if err != nil {
															log.Println("Deep actions new tx undelegate contract transaction error 1:", err)
														}
													} else {
														log.Println("Deep actions new tx undelegate contract transaction error 2")
													}
													break
												}
											}
											break
										}
									case 2:
										{
											if t.Comment.Title == "delegate_reward_transaction" && t.To == config.DelegateScAddress {
												timestampUnix := apparel.UnixFromStringTimestamp(t.Timestamp)
												_ = delegate_con.Bonus(t.Timestamp, timestampUnix)
											}
											break
										}
									case 5:
										{
											switch t.Comment.Title {
											case "undelegate_contract_transaction":
												{
													if t.From == config.DelegateScAddress {
														timestampUnix := apparel.UnixFromStringTimestamp(t.Timestamp)
														err := delegate_con.UnDelegate(t.To, t.Amount, timestampUnix)
														if err != nil {
															log.Println("Deep actions new tx undelegate contract transaction error 1:", err)
														}
													} else {
														log.Println("Deep actions new tx undelegate contract transaction error 2")
													}
													break
												}
											}
											break
										}
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

						storage.Update()

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
	timestamp := apparel.Timestamp()
	transaction := *deep_actions.NewTx(
		2,
		apparel.GetNonce(timestamp),
		"",
		config.BlockHeight,
		config.GenesisAddress,
		config.NodeUwAddress,
		storage.CalculateReward(config.NodeNdAddress),
		config.RewardTokenLabel,
		timestamp,
		0,
		crypt.SignMessageWithSecretKey(config.GenesisSecretKey, []byte(config.GenesisAddress)),
		*deep_actions.NewComment(
			"reward_transaction",
			nil,
		),
	)

	jsonForHash, err := json.Marshal(transaction)
	if err != nil {
		log.Println("Reward transaction error:", err)
	}

	transaction.HashTx = crypt.GetHash(jsonForHash)

	return transaction
}

func delegateTransaction() deep_actions.Tx {
	timestamp := apparel.Timestamp()
	transaction := *deep_actions.NewTx(
		2,
		apparel.GetNonce(timestamp),
		"",
		config.BlockHeight,
		config.GenesisAddress,
		config.DelegateScAddress,
		storage.CalculateReward(config.DelegateScAddress),
		config.RewardTokenLabel,
		timestamp,
		0,
		crypt.SignMessageWithSecretKey(config.GenesisSecretKey, []byte(config.GenesisAddress)),
		*deep_actions.NewComment(
			"delegate_reward_transaction",
			nil,
		),
	)

	jsonForHash, err := json.Marshal(transaction)
	if err != nil {
		log.Println("Delegate transaction error:", err)
	}

	transaction.HashTx = crypt.GetHash(jsonForHash)

	return transaction
}

func addNodesForVote() []deep_actions.Vote {
	var votes []deep_actions.Vote

	if memory.ValidatorsMemory != nil {
		for _, node := range memory.ValidatorsMemory {
			votes = append(votes, *deep_actions.NewVote(node.Address, nil, config.BlockHeight, false))
		}
	}

	return votes
}

func validateBlock(block storage.Block) bool {
	chain := deep_actions.Chain{}
	chain.Header.TxCounter = int64(len(block.Body))
	chain.Header.ProposerSignature = block.ProposerSignature
	chain.Header.Proposer = block.Proposer
	chain.Header.PrevHash = block.PrevHash

	for _, t := range block.Body {
		jsonString, _ := json.Marshal(t)
		t.HashTx = crypt.GetHash(jsonString)
		chain.Txs = append(chain.Txs, t)
	}

	if err := validation.ValidateBlock(chain); err != nil {
		fmt.Println(err)
		return false
	} else {
		return true
	}
}
