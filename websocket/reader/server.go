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
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/storage/deep_actions"
	"node/storage/validation"
	websocket2 "node/websocket"
	"node/websocket/sender"
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
		log.Println("REFUND TRANSACTION")
	}

	storage.TransactionsMemory = append(storage.TransactionsMemory, tx)
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
				var blocks deep_actions.Chains

				minHeight := request.SenderHeight
				maxHeight := config.BlockHeight
				if maxHeight-minHeight > 500 {
					maxHeight = minHeight + 500
				}

				for i := minHeight; i < maxHeight; i++ {
					c := deep_actions.Chain{}
					err := json.Unmarshal([]byte(deep_actions.GetChainJson(i)), &c)
					if err != nil {
						log.Println("Websocket server reader download blocks error 2:", err)
					} else {
						if c.Hash != "" {
							blocks = append(blocks, c)
						}
					}
				}

				body, err := json.Marshal(blocks)
				if err != nil {
					log.Println("Websocket server reader download blocks error 3:", err)
				}

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
			var chains deep_actions.Chains
			_ = json.Unmarshal([]byte(body), &chains)
			storage.NewBlocksForStart(chains)

			for _, chain := range chains {
				for _, t := range chain.Txs {
					commentData := make(map[string]interface{})
					_ = json.Unmarshal(t.Comment.Data, &commentData)
					switch t.Type {
					case 1:
						blockchain.ExecutionSmartContractsWithType1Transaction(t)
						break
					case 2:
						blockchain.ExecutionSmartContractsWithType2Transaction(t)
						break
					case 3:
						blockchain.ExecutionSmartContractsWithType3Transaction(t)
						break
					case 5:
						blockchain.ExecutionSmartContractsWithType5Transaction(t)
						break
					}
				}
			}
			log.Println("Height after receiving blocks:", config.BlockHeight)
		}
	}
}
