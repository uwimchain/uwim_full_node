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
)

type Client struct {
	Address    string  `json:"address"`
	Balance    float64 `json:"balance"`
	UpdateTime int64   `json:"update_time"`
}

// Функция делигирования
func Delegate(address string, amount float64, timestamp int64) error {
	err := updateBalance(address, amount, true, timestamp)
	if err != nil {
		return err
	}

	return nil
}

// Функция начисления бонусов за вложение на аакаунты в зависимости от их баланса
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
			//timestamp := apparel.Timestamp()

			for _, client := range clients {
				if client.Balance >= 10000 {
					amount, _ := apparel.Round(client.Balance * (0.12 / 30 / (60 * 60 * 24 / 51 / 6)))
					_ = updateBalance(client.Address, amount, true, timestampUnix)
					contracts.StorageUpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
				} else if client.Balance >= 1000 {
					amount, _ := apparel.Round(client.Balance * (0.08 / 30 / (60 * 60 * 24 / 51 / 6)))
					_ = updateBalance(client.Address, amount, true, timestampUnix)
					contracts.StorageUpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
				} else if client.Balance >= 100 {
					amount, _ := apparel.Round(client.Balance * (0.05 / 30 / (60 * 60 * 24 / 51 / 6)))
					_ = updateBalance(client.Address, amount, true, timestampUnix)
					contracts.StorageUpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
				}
			}
		}
	}

	return nil
}

// Функция разделегирования токенов пользователя
func SendUnDelegate(address string, amount float64) error {
	amount, _ = apparel.Round(amount)
	if memory.IsNodeProposer() {
		client := getClient(address)
		if client.Balance <= 0 {
			return errors.New("Blockchain contracts delegate contract undelegate error 1: not coins for undelegate")
		}

		to := address
		timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)
		nonce := apparel.GetNonce(timestampD)
		sign := crypt.SignMessageWithSecretKey(config.NodeSecretKey, []byte(address))

		if amount <= 0 || amount >= client.Balance {
			amount = client.Balance
		}

		comment := *contracts.NewComment("undelegate_contract_transaction", nil)

		tx := contracts.NewTx(
			5,
			0,
			"",
			0,
			config.DelegateScAddress,
			to,
			amount,
			config.DelegateToken,
			"",
			apparel.CalcTax(amount),
			nil,
			comment,
		)

		jsonForHash, err := json.Marshal(tx)
		if err != nil {
			return errors.New("Blockchain contracts delegate contract undelegate error 3")
		}

		commentData, err := json.Marshal(contracts.NewContractCommentData(config.NodeNdAddress, jsonForHash))
		if err != nil {
			return errors.New("Blockchain contracts delegate contract undelegate error 4")
		}

		tx.Nonce = nonce
		tx.Height = config.BlockHeight
		tx.Timestamp = timestampD
		tx.Signature = sign
		tx.Comment.Data = commentData

		jsonString, _ := json.Marshal(contracts.Tx{
			Type:       tx.Type,
			Nonce:      tx.Nonce,
			From:       tx.From,
			To:         tx.To,
			Amount:     tx.Amount,
			TokenLabel: tx.TokenLabel,
			Comment:    tx.Comment,
		})
		tx.Signature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)

		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *tx)
		contracts.SendTx(*tx)

		// TxDB.Put(strconv.FormatInt(nonce, 10), string(jsonString))
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

// Функция разделегирования токенов пользователя
func UnDelegate(address string, amount float64, timestamp int64) error {
	amount, _ = apparel.Round(amount)
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

// Функция получения баланса делегирования пользователя
func GetBalance(address string) Client {
	balance := getClient(address)
	return balance
}

// вспомогательная функция обновления баланса пользователя при делегировании
// и разделегировании
func updateBalance(address string, amount float64, side bool, timestamp int64) error {
	amount, _ = apparel.Round(amount)
	client := getClient(address)

	amount, _ = apparel.Round(amount)
	client.Balance, _ = apparel.Round(client.Balance)

	switch side {
	case true:
		client.Balance += amount
		break
	case false:
		if client.Balance < amount {
			return errors.New("Blockchain contracts delegate contract update balance error 1")
		} else {
			client.Balance -= amount
		}
	}

	// Сохранение изменённого баланса пользователя в базу данных
	jsonString, err := json.Marshal(Client{
		Address:    address,
		Balance:    client.Balance,
		UpdateTime: timestamp,
	})
	if err != nil {
		return errors.New("Blockchain contracts delegate contract update balance error 2")
	}

	ClientDB.Put(address, string(jsonString))
	return nil
}

// вспомогательная функция для получения информациии о клиенте смарт-контракта по его адресу
func getClient(address string) Client {
	row := ClientDB.Get(address).Value
	client := Client{}
	_ = json.Unmarshal([]byte(row), &client)
	return client
}

func UnDelegateValidate(address string, amount float64, checkSum []byte) error {
	amount, _ = apparel.Round(amount)
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
		apparel.CalcTax(amount),
		nil,
		comment,
	)

	jsonForHash, _ := json.Marshal(t)

	if string(jsonForHash) != string(checkSum) {
		return errors.New("undelegate transaction checksum error")
	}

	return nil
}
