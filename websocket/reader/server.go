package reader

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"net/http"
	"node/apparel"
	"node/blockchain"
	"node/blockchain/contracts/delegate_con"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/storage/validation"
	websocket_my "node/websocket"
	"node/websocket/sender"
	"os"
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
			newBlock(body)
			break
		case "BlockVote":
			blockVote(body)
			break
		case "NewTx":
			newTx(body)
			break
		case "GetVersion":
			if err := getVersion(body); err != nil {
				log.Println(err)
			}

			break
		case "DownloadBlocks":
			downloadBlocks(body)
			break
		default:
			log.Println("Websocket reader server request reader error 3: incorrect request data type")
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

func newBlock(body string) {
	if memory.IsValidator() {
		if blockchain.NodeOperationMemory.Operation == 1 {
			if !memory.IsNodeProposer() {
				block := storage.Block{}
				err := json.Unmarshal([]byte(body), &block)
				if err != nil {
					log.Println("Websocket server reader new block error:", err)
				}
				storage.BlockMemory = block
				log.Println("BLock height:", config.BlockHeight)
			}

			if storage.BlockMemory.Height != config.BlockHeight {
				if storage.BlockMemory.Height > config.BlockHeight {
					sender.DownloadBlocksFromNodes()
				}
			}
		}
	}
}

func blockVote(body string) {
	vote := deep_actions.Vote{}
	err := json.Unmarshal([]byte(body), &vote)
	if err != nil {
		log.Println("Websocket server reader block vote error:", err)
	}
	publicKey, _ := crypt.PublicKeyFromAddress(vote.Proposer)
	if vote.BlockHeight == config.BlockHeight && crypt.VerifySign(publicKey, []byte(vote.Proposer), vote.Signature) {
		for i := range storage.BlockMemory.Votes {
			if storage.BlockMemory.Votes[i].Proposer == vote.Proposer {
				storage.BlockMemory.Votes[i].Vote = vote.Vote
				storage.BlockMemory.Votes[i].BlockHeight = vote.BlockHeight
				storage.BlockMemory.Votes[i].Signature = vote.Signature
				break
			}
		}
	}
}

func newTx(body string) {
	t := deep_actions.Tx{}
	err := json.Unmarshal([]byte(body), &t)
	if err != nil {
		log.Println("Websocket server reader new tx error 1:", err)
	} else {
		if err := validateTx(t); err != nil {
			log.Println("Websocket server reader new tx error 2:", err)
		} else {
			storage.TransactionsMemory = append(storage.TransactionsMemory, t)
		}
	}
}

func getVersion(body string) error {
	r := websocket_my.RequestSign{}
	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		return errors.New(fmt.Sprintf("Websocket server reader get config error 1: %v", err))
	}

	pubicKey, _ := crypt.PublicKeyFromAddress(r.Address)
	if !crypt.VerifySign(pubicKey, []byte(r.Address), r.Sign) {
		return errors.New(fmt.Sprintf("Websocket server reader new tx error 2: %v", err))
	}

	if r.Data != nil {
		if string(r.Data) != config.Version {
			log.Println("Version error")
			os.Exit(1)
		}
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
											timestampUnix := apparel.UnixFromStringTimestamp(t.Timestamp)
											err := delegate_con.Delegate(t.From, t.Amount, timestampUnix)
											if err != nil {
												log.Println("Deep actions download new tx delegate contract transaction error 1:", err)
											}
										} else {
											log.Println("Deep actions download new tx delegate contract transaction error 2")
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

				// Создание смарт-контрактов при подкачке
				//for _, b := range blocks {
				//	for _, t := range b.Txs {
				//		if crypt.IsAddressSmartContract(t.To) && t.Type == 4 {
				//			contract := deep_actions.Contract{}
				//			err := json.Unmarshal(t.Comment.Data, &contract)
				//			if err != nil {
				//				log.Println(err)
				//			} else {
				//				testCon := test_con.TestContract{}
				//				err := json.Unmarshal(contract.Data, &testCon)
				//				if err != nil {
				//					log.Println(err)
				//				} else {
				//					err := test_con.AddTestSmartContract(t.To, testCon.To, t.TokenLabel)
				//					if err != nil {
				//						log.Println(err)
				//					}
				//				}
				//			}
				//		}
				//	}
				//}
			}
			log.Println("Height after receiving blocks:", config.BlockHeight)
		}
	}
}
