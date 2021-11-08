package delegate_con

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
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

type DelegateArgs struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

func NewDelegateArgs(address string, amount float64) (*DelegateArgs, error) {
	return &DelegateArgs{Address: address, Amount: amount}, nil
}

func Delegate(args *DelegateArgs) error {
	timestamp := apparel.TimestampUnix()
	err := updateBalance(args.Address, args.Amount, true, timestamp)
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
				address := contracts.GetAddress(config.DelegateScAddress)

				if config.BlockHeight < config.AnnualBlockHeight {
					if client.Balance >= 10000 {
						amount := apparel.Round(client.Balance * (0.12 / 30 / (60 * 60 * 24 / 51 / 6)))
						_ = updateBalance(client.Address, amount, true, timestampUnix)
						address.UpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
					} else if client.Balance >= 1000 {
						amount := apparel.Round(client.Balance * (0.08 / 30 / (60 * 60 * 24 / 51 / 6)))
						_ = updateBalance(client.Address, amount, true, timestampUnix)
						address.UpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
					} else if client.Balance >= 100 {
						amount := apparel.Round(client.Balance * (0.05 / 30 / (60 * 60 * 24 / 51 / 6)))
						_ = updateBalance(client.Address, amount, true, timestampUnix)
						address.UpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
					}
				} else {
					if client.Balance >= 100 {
						emitRateIdx := config.GetEmitRateIdx()
						if emitRateIdx >= 0 {
							amount := apparel.Round(((client.Balance * config.EmitRate[emitRateIdx]) / 1000000) * float64(len(memory.ValidatorsMemory)) * config.DelegateEmitRate)
							_ = updateBalance(client.Address, amount, true, timestampUnix)
							address.UpdateBalance(config.DelegateScAddress, *contracts.NewBalance(config.DelegateToken, amount, timestamp), false)
						}
					}
				}
			}
		}
	}

	return nil
}

func SendUnDelegate(args *DelegateArgs) error {
	args.Amount = apparel.Round(args.Amount)
	client := getClient(args.Address)
	if client.Balance <= 0 {
		return errors.New("Blockchain contracts delegate contract undelegate error 1: not coins for undelegate")
	}

	timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)
	if args.Amount <= 0 || args.Amount >= client.Balance {
		args.Amount = client.Balance
	}

	txCommentSign := contracts.NewBuyTokenSign(
		config.NodeNdAddress,
	)

	contracts.SendNewScTx(timestampD, config.BlockHeight, config.DelegateScAddress, args.Address, args.Amount,
		config.DelegateToken, "undelegate_contract_transaction", txCommentSign)
	return nil
}

func UnDelegate(args *DelegateArgs) error {
	timestamp := apparel.TimestampUnix()
	args.Amount = apparel.Round(args.Amount)
	client := getClient(args.Address)
	log.Println(args.Address)
	log.Println(client)

	if client.Balance <= 0 {
		return errors.New("Blockchain contracts delegate contract undelegate error 1: not coins for undelegate")
	}

	err := updateBalance(args.Address, args.Amount, false, timestamp)
	if err != nil {
		return err
	}

	return nil
}

func GetBalance(address string) Client {
	balance := getClient(address)
	return balance
}

func updateBalance(address string, amount float64, side bool, timestamp int64) error {
	//amount, _ = apparel.Round(amount)
	amount = apparel.Round(amount)
	client := getClient(address)

	//amount, _ = apparel.Round(amount)
	amount = apparel.Round(amount)
	//client.Balance, _ = apparel.Round(client.Balance)
	client.Balance = apparel.Round(client.Balance)

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

func getClient(address string) Client {
	row := ClientDB.Get(address).Value
	client := Client{}
	_ = json.Unmarshal([]byte(row), &client)
	return client
}
