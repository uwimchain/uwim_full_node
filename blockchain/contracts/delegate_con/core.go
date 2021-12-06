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
	Address    string           `json:"address"`
	Balance    float64          `json:"balance"`
	UpdateTime contracts.String `json:"update_time"`
}

type DelegateArgs struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

func NewDelegateArgs(address string, amount float64) (*DelegateArgs, error) {
	return &DelegateArgs{Address: address, Amount: amount}, nil
}

func Delegate(args *DelegateArgs) error {
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
	err := updateBalance(args.Address, args.Amount, true, timestamp)
	if err != nil {
		return err
	}

	return nil
}

func Bonus(timestamp string) error {
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
						amount := apparel.Round(client.Balance * (0.12 / 30 / (60 * 60 * 24 / 51 / float64(len(memory.ValidatorsMemory)))))
						_ = updateBalance(client.Address, amount, true, timestamp)
						address.UpdateBalance(config.DelegateScAddress, amount, config.DelegateToken, timestamp, false)
					} else if client.Balance >= 1000 {
						amount := apparel.Round(client.Balance * (0.08 / 30 / (60 * 60 * 24 / 51 / float64(len(memory.ValidatorsMemory)))))
						_ = updateBalance(client.Address, amount, true, timestamp)
						address.UpdateBalance(config.DelegateScAddress, amount, config.DelegateToken, timestamp, false)
					} else if client.Balance >= 100 {
						amount := apparel.Round(client.Balance * (0.05 / 30 / (60 * 60 * 24 / 51 / float64(len(memory.ValidatorsMemory)))))
						_ = updateBalance(client.Address, amount, true, timestamp)
						address.UpdateBalance(config.DelegateScAddress, amount, config.DelegateToken, timestamp, false)
					}
				} else {
					if client.Balance >= 100 {
						emitRateIdx := config.GetEmitRateIdx()
						if emitRateIdx >= 0 {
							amount := apparel.Round(((client.Balance * config.EmitRate[emitRateIdx]) / 1000000) * float64(len(memory.ValidatorsMemory)) * config.DelegateEmitRate)
							_ = updateBalance(client.Address, amount, true, timestamp)
							address.UpdateBalance(config.DelegateScAddress, amount, config.DelegateToken, timestamp, false)
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

	if args.Amount <= 0 || args.Amount >= client.Balance {
		args.Amount = client.Balance
	}

	contracts.SendNewScTx(config.DelegateScAddress, args.Address, args.Amount, config.DelegateToken, "undelegate_contract_transaction")
	return nil
}

func UnDelegate(args *DelegateArgs) error {
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
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

func updateBalance(address string, amount float64, side bool, timestamp string) error {
	amount = apparel.Round(amount)
	client := getClient(address)

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

	jsonString, err := json.Marshal(Client{
		Address:    address,
		Balance:    client.Balance,
		UpdateTime: contracts.String(timestamp),
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
