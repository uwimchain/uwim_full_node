package delegate_con

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/memory"
	"strconv"
)

var (
	db = contracts.Database{}

	ClientDB = db.NewConnection("blockchain/contracts/delegate_con/storage/contract_client")
	txDB     = db.NewConnection("blockchain/contracts/delegate_con/storage/contract_txs")
)

type Client struct {
	Address    string  `json:"address"`
	Balance    float64 `json:"balance"`
	UpdateTime int64   `json:"update_time"`
}

func NewClient(address string, balance float64, updateTime int64) *Client {
	return &Client{Address: address, Balance: balance, UpdateTime: updateTime}
}

func Delegate(address string, amount float64, timestamp int64) error {
	err := updateBalance(address, amount, true, timestamp)
	if err != nil {
		return err
	}

	return nil
}

func Bonus(timestamp string, timestampUnix int64) error {
	rows := ClientDB.GetAll("")
	if rows != nil {
		var clients []Client

		for _, row := range rows {
			client := Client{}
			err := json.Unmarshal([]byte(row.Value), &client)
			if err != nil {
				return err
			} else {
				clients = append(clients, client)
			}
		}

		if clients != nil {

			for _, client := range clients {
				if client.Balance >= 10000 {
					amount := client.Balance * (0.12 / 30 / (60 * 60 * 24 / 51 / 6))
					_ = updateBalance(client.Address, amount, true, timestampUnix)
					contracts.StorageUpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
				} else if client.Balance >= 1000 {
					amount := client.Balance * (0.08 / 30 / (60 * 60 * 24 / 51 / 6))
					_ = updateBalance(client.Address, amount, true, timestampUnix)
					contracts.StorageUpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
				} else if client.Balance >= 100 {
					amount := client.Balance * (0.05 / 30 / (60 * 60 * 24 / 51 / 6))
					_ = updateBalance(client.Address, amount, true, timestampUnix)
					contracts.StorageUpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
				}
			}
		}
	}

	return nil
}

func SendUnDelegate(address string, amount float64) error {
	if memory.IsNodeProposer() {
		client := getClient(address)
		if client.Balance <= 0 {
			return errors.New("Blockchain contracts delegate contract undelegate error 1: not coins for undelegate")
		}

		to := address
		timestamp := apparel.Timestamp()
		nonce := apparel.GetNonce(timestamp)
		sign := crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(address))

		if amount <= 0 || amount >= client.Balance {
			amount = client.Balance
		}

		comment := *contracts.NewComment("undelegate_contract_transaction", nil)

		t := contracts.NewTx(
			5,
			0,
			"",
			0,
			config.DelegateScAddress,
			to,
			amount,
			config.DelegateToken,
			"",
			apparel.CalcTax(amount*config.TaxConversion*config.Tax),
			nil,
			comment,
		)

		jsonForHash, err := json.Marshal(t)
		if err != nil {
			return errors.New("Blockchain contracts delegate contract undelegate error 3")
		}

		commentData, err := json.Marshal(contracts.NewContractCommentData(config.NodeNdAddress, jsonForHash))
		if err != nil {
			return errors.New("Blockchain contracts delegate contract undelegate error 4")
		}

		t.Nonce = nonce
		t.Height = config.BlockHeight
		t.Timestamp = timestamp
		t.Signature = sign
		t.Comment.Data = commentData

		jsonString, err := json.Marshal(t)
		if err != nil {
			return errors.New("Blockchain contracts delegate contract undelegate error 5")
		}

		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *t)
		contracts.SendTx(jsonString)
		newTxs(nonce, jsonString)
	}

	return nil
}

type UndelegateCommentData struct {
	Amount float64 `json:"amount"`
}

func NewUndelegateCommentData(amount float64) *UndelegateCommentData {
	return &UndelegateCommentData{
		Amount: amount,
	}
}

func UnDelegate(address string, amount float64, timestamp int64) error {
	client := getClient(address)
	if client.Balance <= 0 {
		return errors.New("Blockchain contracts delegate contract undelegate error 1: not coins for undelegate")
	}

	err := updateBalance(address, amount, false, timestamp)
	if err != nil {
		return err
	}

	return nil
}

func GetBalance(address string) Client {
	balance := getClient(address)
	return balance
}

func newTxs(nonce int64, jsonString []byte) {
	txDB.Put(strconv.FormatInt(nonce, 10), string(jsonString))
}

func updateBalance(address string, amount float64, side bool, timestamp int64) error {
	client := getClient(address)

	if side {
		client.Balance += amount
	} else {
		if client.Balance < amount {
			return errors.New("Blockchain contracts delegate contract update balance error 1")
		} else {
			client.Balance -= amount
		}
	}

	jsonString, err := json.Marshal(NewClient(address, client.Balance, timestamp))
	if err != nil {
		return errors.New("Blockchain contracts delegate contract update balance error 2")
	}

	ClientDB.Put(address, string(jsonString))
	return nil
}

func getClient(address string) Client {
	row := ClientDB.Get(address).Value
	client := Client{}
	_ = json.Unmarshal([]byte(row), &client)
	return client
}

func UnDelegateValidate(address string, amount float64, checkSum []byte) error {
	to := address
	comment := *contracts.NewComment("undelegate_contract_transaction", nil)
	t := contracts.NewTx(
		5,
		0,
		"",
		0,
		config.DelegateScAddress,
		to,
		amount,
		config.DelegateToken,
		"",
		apparel.CalcTax(amount*config.TaxConversion*config.Tax),
		nil,
		comment,
	)

	jsonForHash, _ := json.Marshal(t)

	if string(jsonForHash) != string(checkSum) {
		return errors.New("undelegate transaction checksum error")
	}

	return nil
}
