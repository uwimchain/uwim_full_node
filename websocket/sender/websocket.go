package sender

import (
	"encoding/json"
	"fmt"
	"log"
	"node/config"
	"node/crypt"
	"node/memory"
	"node/storage"
	"node/websocket"
)

func SendTx(tx []byte) {
	message, _ := json.Marshal(websocket.RequestSign{
		SenderIp: config.Ip,
		Sign:     crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
		Address:  config.NodeNdAddress,
		Data:     tx,
	})

	requestsSender("NewTx", message)
}

func SendVersion(version []byte) {
	message, _ := json.Marshal(websocket.RequestSign{
		SenderIp: config.Ip,
		Sign:     crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(config.NodeNdAddress)),
		Address:  config.NodeNdAddress,
		Data:     version,
	})

	log.Println("Check Version")
	requestsSender("GetVersion", message)
}

func SendBlockVote(vote []byte) {
	requestsSender("BlockVote", vote)
}

func SendNewBlock() {
	body, err := json.Marshal(storage.BlockMemory)
	if err != nil {
		log.Println("Send New Block error:", err)
	}

	requestsSender("NewBlock", body)
}

type GetBlocksRequest struct {
	SenderIp     string `json:"senderIp"`
	SenderHeight int64  `json:"senderHeight"`
}

func NewGetBlocksRequest(senderIp string, senderHeight int64) *GetBlocksRequest {
	return &GetBlocksRequest{SenderIp: senderIp, SenderHeight: senderHeight}
}

func DownloadBlocksFromNodes() {

	body, err := json.Marshal(NewGetBlocksRequest(config.Ip, config.BlockHeight))
	if err != nil {
		log.Println("Download Blocks From Nodes error:", err)
	}

	for _, validator := range memory.ValidatorsMemory {
		if validator.Ip != config.Ip {
			if err := Client(validator.Ip, Request{
				DataType: "DownloadBlocks",
				Body:     string(body),
			}, "/ws"); err != nil {
				log.Println("Client Error: " + fmt.Sprintf("%v", err))
			} else {
				return
			}
		}
	}
}

func requestsSender(dataType string, body []byte) {
	if len(memory.ValidatorsMemory) != 0 {
		for _, node := range memory.ValidatorsMemory {
			if node.Address != config.NodeNdAddress && node.Ip != config.Ip {
				if err := Client(node.Ip, Request{
					DataType: dataType,
					Body:     string(body),
				}, "/ws"); err != nil {
					log.Println("Client Error: " + fmt.Sprintf("%v", err))
				}
			}
		}
	} else {
		log.Println("Client Error: NodesMemory array is empty.")
	}
}
