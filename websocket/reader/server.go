package reader

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"net/http"
	"node/blockchain"
	"node/blockchain/contracts/business_token_con"
	"node/blockchain/contracts/delegate_con"
	"node/blockchain/contracts/donate_token_con"
	"node/blockchain/contracts/my_token_con"
	"node/blockchain/contracts/trade_token_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/metrics"
	"node/storage"
	"node/storage/deep_actions"
	"node/storage/validation"
	websocket2 "node/websocket"
	"node/websocket/sender"
	"strconv"
)

func Init() {
	http.HandleFunc("/ws", RequestsReader)
	log.Fatal("Websocket reader server init error:", http.ListenAndServe(*flag.String("addr", config.Ip, "http service address"), nil))
}

var upgrader = websocket.Upgrader{}

func RequestsReader(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Websocket reader server request reader error 1:", err)
		return
	}

	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("Websocket reader server request reader error 2:", err)
			break
		}

		dataType, body := decodeMessage(message)
		switch dataType {
		case "NewBlock":
			if err := newBlock(body); err != nil {
				log.Println(err)
			}

			break
		case "BlockVote":
			blockVote(body)
			break
		case "NewTx":
			if err := newTx(body); err != nil {
				log.Println(err)
			}

			break
		case "GetVersion":
			if err := getVersion(body); err != nil {
				log.Println(err)
			}

			break
		case "GetProposer":
			var requestSign websocket2.RequestSign
			err := json.Unmarshal([]byte(body), &requestSign)
			if err != nil {
				log.Println("Websocket reader server request reader error 3: ", err)
				break
			}

			sender.SendProposer(requestSign.SenderIp)

			break
		case "Proposer":
			var requestSign websocket2.RequestSign
			err := json.Unmarshal([]byte(body), &requestSign)
			if err != nil {
				log.Println("Websocket reader server request reader error 4: ", err)
				break
			}

			storage.BlockMemory = storage.Block{}
			if !crypt.IsAddressNode(string(requestSign.Data)) || !memory.IsNodeValidator(string(requestSign.Data)) {
				memory.Proposer = string(requestSign.Data)
			}

			break
		case "DownloadBlocks":
			downloadBlocks(body)

			break
		default:
			log.Println("Websocket reader server request reader error 5: incorrect request data type")

			break
		}
		break
	}
}

func decodeMessage(message []byte) (string, string) {
	request := sender.Request{}
	err := json.Unmarshal(message, &request)
	if err != nil {
		log.Println("Websocket reader server decode message error:", err)
	}
	return request.DataType, request.Body
}

func newBlock(body string) error {
	if !memory.IsValidator() {
		return errors.New("Websocket server reader new block error 1: this node is not validator")
	}

	if blockchain.NodeOperationMemory.Operation != 1 {
		return errors.New("Websocket server reader new block error 2: blockchain operation memory error")
	}

	//if memory.IsNodeProposer() {
	//	return errors.New("Websocket server reader new block error 3: this node is proposer")
	//}

	r := websocket2.RequestSign{}
	b := storage.Block{}

	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		return errors.New(fmt.Sprintf("Websocket server reader new block error 4: %v", err))
	}

	pubicKey, _ := crypt.PublicKeyFromAddress(r.Address)
	if !crypt.VerifySign(pubicKey, []byte(r.Address), r.Sign) {
		return errors.New(fmt.Sprintf("Websocket server reader new block error 5: %v", err))
	}

	if err := json.Unmarshal(r.Data, &b); err != nil {
		return errors.New(fmt.Sprintf("Websocket server reader new block error 6: %v", err))
	}

	if b.Proposer != memory.Proposer {

		sender.GetProposer()
		return errors.New("Websocket server reader new block error 3: this node is not a proposer")
	}

	storage.BlockMemory = b
	log.Println("BLock height:", config.BlockHeight)

	if storage.BlockMemory.Height != config.BlockHeight {
		if storage.BlockMemory.Height > config.BlockHeight {
			sender.DownloadBlocksFromNodes()
		}
	}

	return nil
}

func blockVote(body string) {
	vote := deep_actions.Vote{}
	err := json.Unmarshal([]byte(body), &vote)
	if err != nil {
		log.Println("Websocket server reader block vote error:", err)
	}
	publicKey, _ := crypt.PublicKeyFromAddress(vote.Proposer)

	jsonString, _ := json.Marshal(deep_actions.Vote{
		Proposer:    vote.Proposer,
		Signature:   nil,
		BlockHeight: vote.BlockHeight,
		Vote:        vote.Vote,
	})

	if vote.BlockHeight == config.BlockHeight && crypt.VerifySign(publicKey, jsonString, vote.Signature) {
		for i := range storage.BlockMemory.Votes {
			if storage.BlockMemory.Votes[i].Proposer == vote.Proposer {
				//storage.BlockMemory.Votes[i].Vote = vote.Vote
				//storage.BlockMemory.Votes[i].BlockHeight = vote.BlockHeight
				//storage.BlockMemory.Votes[i].Signature = vote.Signature
				storage.BlockMemory.Votes[i] = vote
				break
			}
		}
	}
}

func newTx(body string) error {
	tx := deep_actions.Tx{}
	r := websocket2.RequestSign{}

	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		return errors.New(fmt.Sprintf("Websocket server reader new tx error 1: %v", err))
	}

	pubicKey, _ := crypt.PublicKeyFromAddress(r.Address)
	if !crypt.VerifySign(pubicKey, []byte(r.Address), r.Sign) {
		return errors.New(fmt.Sprintf("Websocket server reader new tx error 2: %v", err))
	}

	if err := json.Unmarshal(r.Data, &tx); err != nil {
		return errors.New(fmt.Sprintf("Websocket server reader new tx error 3: %v", err))
	}

	if err := validateTx(tx); err != nil {
		return errors.New(fmt.Sprintf("Websocket server reader new tx error 4: %v", err))
	}

	if tx.Comment.Title == "refund_transaction" {
		log.Println("GG REFUND TRANSACTION")
	}

	storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
	//if !memory.IsNodeValidator(r.SenderIp, r.Address) {
	if !memory.IsNodeValidator(r.Address) {
		sender.SendTx(tx)
	}

	return nil
}

func getVersion(body string) error {
	r := websocket2.RequestSign{}
	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		return errors.New(fmt.Sprintf("Websocket server reader get config error 1: %v", err))
	}

	pubicKey, _ := crypt.PublicKeyFromAddress(r.Address)
	if !crypt.VerifySign(pubicKey, []byte(r.Address), r.Sign) {
		return errors.New(fmt.Sprintf("Websocket server reader new tx error 2: %v", err))
	}

	if r.Data == nil {
		sender.SendVersion([]byte(config.Version), r.SenderIp)
	}

	return nil
}

func validateTx(transaction deep_actions.Tx) error {
	transactions := storage.TransactionsMemory
	transactions = append(transactions, transaction)

	return validation.ValidateTxs(transactions)
}

func downloadBlocks(body string) {
	request := sender.GetBlocksRequest{}

	err := json.Unmarshal([]byte(body), &request)
	if err != nil {
		log.Println("Websocket server reader download blocks error 1:", err)
	}

	if request.SenderIp != "" {
		if memory.IsValidator() {
			if config.BlockHeight != request.SenderHeight && config.BlockHeight > request.SenderHeight {
				log.Println("Send blocks for:", request.SenderIp)
				var blocks []deep_actions.Chain

				// Если высота ноды, желающей получить блоки меньше высоты ноды больше, чем на 500, то нодаотправит ей максимум 500 блоков
				// иначе, отправит ей разницу между высотами нод
				minHeight := request.SenderHeight
				maxHeight := config.BlockHeight
				if maxHeight-minHeight > 500 {
					maxHeight = minHeight + 500
				}

				// Выборка нужного количества блоков из базы данных по высоте
				for i := minHeight; i < maxHeight; i++ {
					c := deep_actions.Chain{}
					err := json.Unmarshal([]byte(storage.GetBLockForHeight(i)), &c)
					if err != nil {
						log.Println("Websocket server reader download blocks error 2:", err)
					} else {
						if c.Hash != "" {
							blocks = append(blocks, c)
						}
					}
				}

				// Сборка полученых из бд блоков в JSON для отправки
				body, err := json.Marshal(blocks)
				if err != nil {
					log.Println("Websocket server reader download blocks error 3:", err)
				}

				// Отправка собранных в JSON блоков для ноды, которая запросила подкачку
				if err := sender.Client(request.SenderIp, sender.Request{
					DataType: "DownloadBlocks",
					Body:     string(body),
				}, "/ws"); err != nil {
					log.Println("Websocket server reader download blocks error 3:", err)
				}
			}
		}
	} else {
		if !memory.DownloadBlocks {
			memory.DownloadBlocks = true
			log.Println("Get blocks")
			var blocks []deep_actions.Chain
			err := json.Unmarshal([]byte(body), &blocks)

			if err != nil {
				log.Println("Websocket server reader download blocks error 4:", err)
			} else {
				// Запись блоков в базу данных при подкачке
				storage.NewBlocksForStart(blocks)

				for _, block := range blocks {
					for _, t := range block.Txs {
						switch t.Type {
						case 1:
							{
								switch t.Comment.Title {
								case "delegate_contract_transaction":
									{
										if t.To == config.DelegateScAddress {
											timestamp, _ := strconv.ParseInt(t.Timestamp, 10, 64)
											err := delegate_con.Delegate(t.From, t.Amount, timestamp)
											if err != nil {
												log.Println("Deep actions download new tx delegate contract transaction error 1:", err)
											}
										} else {
											log.Println("Deep actions download new tx delegate contract transaction error 2")
										}
										break
									}
								case "my_token_contract_confirmation_transaction":
									{
										confirmationArgs, err := my_token_con.NewConfirmationArgs(t.To, t.From, t.Height, t.HashTx)
										if err != nil {
											log.Println("Deep actions new tx confirmation transaction error 1:", err)
											break
										}

										err = my_token_con.Confirmation(confirmationArgs)
										if err != nil {
											log.Println("Deep actions new tx confirmation transaction error 2:", err)
										}

										break
									}
								case "my_token_contract_get_percent_transaction":
									{
										getPercentAgs, err := my_token_con.NewGetPercentArgs(t.To, t.From, t.TokenLabel, t.Height, t.HashTx)
										if err != nil {
											log.Println("Deep actions new tx my contract get percent transaction error 1:", err)
											break
										}

										err = my_token_con.GetPercent(getPercentAgs)
										if err != nil {
											log.Println("Deep actions new tx my contract get percent transaction error 2:", err)
										}

										break
									}
								case "trade_token_contract_add_transaction":
									{
										err := trade_token_con.Add(trade_token_con.NewTradeArgs(t.To, t.From, t.Amount, t.TokenLabel, t.Height, t.HashTx))
										if err != nil {
											log.Println("Deep actions new tx trade contract add transaction error 1:", err)
											break
										}
										break
									}
								case "trade_token_contract_swap_transaction":
									{
										err := trade_token_con.Swap(trade_token_con.NewTradeArgs(t.To, t.From, t.Amount, t.TokenLabel, t.Height, t.HashTx))
										if err != nil {
											log.Println("Deep actions new tx trade contract swap transaction error 1:", err)
											break
										}
										break
									}
								case "trade_token_contract_get_liq_transaction":
									{
										err := trade_token_con.GetLiq(trade_token_con.NewGetArgs(t.To, t.From, t.TokenLabel, t.Height, t.HashTx))
										if err != nil {
											log.Println("Deep actions new tx trade contract get transaction error 1:", err)
											break
										}
										break
									}
								case "trade_token_contract_get_com_transaction":
									{
										err := trade_token_con.GetCom(trade_token_con.NewGetArgs(t.To, t.From, t.TokenLabel, t.Height, t.HashTx))
										if err != nil {
											log.Println("Deep actions new tx trade contract get transaction error 1:", err)
											break
										}
										break
									}
								case "trade_token_contract_fill_config_transaction":
									{
										var scAddressConfig trade_token_con.TradeConfig
										err := json.Unmarshal(t.Comment.Data, &scAddressConfig)
										if err != nil {
											log.Println("Deep actions new tx trade contract fill config transaction error 1:", err)
											break
										}

										err = trade_token_con.FillConfig(trade_token_con.NewFillConfigArgs(t.To, scAddressConfig.Commission))
										if err != nil {
											log.Println("Deep actions new tx trade contract fill config transaction error 2:", err)
											break
										}
										break
									}

								}

								break
							}
						case 2:
							{
								if t.Comment.Title == "delegate_reward_transaction" && t.To == config.DelegateScAddress {
									/*timestampUnix := apparel.UnixFromStringTimestamp(t.Timestamp)*/
									timestamp, _ := strconv.ParseInt(t.Timestamp, 10, 64)
									_ = delegate_con.Bonus(t.Timestamp, timestamp)
								}
								break
							}
						case 3:
							switch t.Comment.Title {
							case "change_token_standard_transaction":
								token := storage.GetAddressToken(t.From)
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
											log.Println("Deep actions new tx change token standard transaction error 5:", err)
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
									_ = my_token_con.ChangeStandard(scAddress)
									break
								case 1:
									err := donate_token_con.ChangeStandard(scAddress)
									if err != nil {
										log.Println("Deep actions new tx change token standard transaction error 7:", err)
										break
									}
									break
								case 4:
									err := business_token_con.ChangeStandard(scAddress)
									if err != nil {
										log.Println("Deep actions new tx change token standard transaction error 8:", err)
										break
									}
									break
								}

								if token.Standard == 5 {
									err := trade_token_con.AddToken(scAddress)
									if err != nil {
										log.Println("Deep actions new tx change token standard transaction error 9:", err)
									}
								}
								break

							case "fill_token_standard_card_transaction":
								token := storage.GetAddressToken(t.From)
								if token.Label == "" {
									break
								}

								switch token.Standard {
								case 4:
									publicKey, err := crypt.PublicKeyFromAddress(t.From)
									if err != nil {
										log.Println("Deep actions new tx fill token card error 4:", err)
										break
									}

									scAddress := crypt.AddressFromPublicKey(metrics.SmartContractPrefix, publicKey)
									err = business_token_con.UpdatePartners(scAddress)
									if err != nil {
										log.Println("Business token contract new tx fill token card error 5:", err)
										break
									}

									break
								}

								break
							}

							break
						case 5:
							{
								switch t.Comment.Title {
								case "undelegate_contract_transaction":
									{
										if t.From == config.DelegateScAddress {
											/*timestampUnix := apparel.UnixFromStringTimestamp(t.Timestamp)*/
											timestamp, _ := strconv.ParseInt(t.Timestamp, 10, 64)
											err := delegate_con.UnDelegate(t.To, t.Amount, timestamp)
											if err != nil {
												log.Println("Deep actions download new tx undelegate contract transaction error 1:", err)
											}
										} else {
											log.Println("Deep actions download new tx undelegate contract transaction error 2")
										}
										break
									}
								}
								break
							}
						}
					}
				}
			}
			log.Println("Height after receiving blocks:", config.BlockHeight)
		}
	}
}
