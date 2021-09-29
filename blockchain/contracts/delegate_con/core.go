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

		timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)

		if amount <= 0 || amount >= client.Balance {
			amount = client.Balance
		}

		txCommentSign, _ := json.Marshal(contracts.NewBuyTokenSign(
			config.NodeNdAddress,
		))

		tx := contracts.NewTx(
			5,
			apparel.GetNonce(timestampD),
			"",
			config.BlockHeight,
			config.DelegateScAddress,
			address,
			amount,
			config.DelegateToken,
			timestampD,
			apparel.CalcTax(amount),
			nil,
			*contracts.NewComment("undelegate_contract_transaction", txCommentSign),
		)

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

		jsonString, _ = json.Marshal(tx)
		tx.HashTx = crypt.GetHash(jsonString)

		contracts.SendTx(*tx)
		*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *tx)
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
