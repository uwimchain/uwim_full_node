package sender

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
)

type Request struct {
	DataType string `json:"dataType"`
	Body     string `json:"body"`
}

func NewRequest(dataType string, body string) *Request {
	return &Request{
		DataType: dataType,
		Body:     body,
	}
}

func Client(address string, data Request, path string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.New(fmt.Sprintf("Client error: %v", err))
	}

	u := url.URL{Scheme: "ws", Host: address, Path: path}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return errors.New(fmt.Sprintf("Client error: %v", err))
	}

	defer c.Close()

	if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		return err
	}
	return nil
}
