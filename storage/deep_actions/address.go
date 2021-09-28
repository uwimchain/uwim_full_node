package deep_actions

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/storage/leveldb"
)

type Address struct {
	Address     string    `json:"address"`
	Balance     []Balance `json:"balance"`
	PublicKey   []byte    `json:"publicKey"`
	FirstTxTime string    `json:"firstTxTime"`
	LastTxTime  string    `json:"lastTxTime"`
	TokenLabel  string    `json:"tokenLabel"`
	ScKeeping   bool      `json:"sc_keeping"`
	Name        string    `json:"name"`
}

type Balance struct {
	TokenLabel string  `json:"tokenLabel"`
	Amount     float64 `json:"amount"`
	UpdateTime string  `json:"updateTime"`
}

func NewBalance(tokenLabel string, amount float64, updateTime string) *Balance {
	return &Balance{TokenLabel: tokenLabel, Amount: amount, UpdateTime: updateTime}
}

func (a *Address) NewAddress(address string, balance []Balance, publicKey []byte, firstTxTime string,
	lastTxTime string) {
	jsonString, err := json.Marshal(Address{
		Address:     address,
		Balance:     balance,
		PublicKey:   publicKey,
		FirstTxTime: firstTxTime,
		LastTxTime:  lastTxTime,
		TokenLabel:  "",
		ScKeeping:   false,
		Name:        "",
	})
	if err != nil {
		log.Println("New Address error: ", err)
	}

	leveldb.AddressDB.Put(address, string(jsonString))
}

func (a *Address) GetAddress(address string) string {
	return leveldb.AddressDB.Get(address).Value
}

func (a *Address) UpdateBalance(address string, balance Balance, side bool) {
	row := a.GetAddress(address)

	if row == "" {
		publicKey, err := crypt.PublicKeyFromAddress(address)
		if err != nil {
			log.Println("Update Balance error 1:", err)
		}

		a.NewAddress(address, nil, publicKey, balance.UpdateTime, balance.UpdateTime)
	}

	row = a.GetAddress(address)
	Addr := Address{}
	err := json.Unmarshal([]byte(row), &Addr)
	if err != nil {
		log.Println("Update Balance error 2:", err)
	}

	Addr.Balance = updateBalance(Addr.Balance, balance, side)

	Addr.LastTxTime = balance.UpdateTime

	jsonString, err := json.Marshal(Addr)
	if err != nil {
		log.Println("Update Balance error 3:", err)
	}

	leveldb.AddressDB.Put(address, string(jsonString))
}

func (a *Address) CheckAddressToken(address string) bool {
	row := a.GetAddress(address)
	if row != "" {
		Addr := Address{}
		err := json.Unmarshal([]byte(row), &Addr)
		if err != nil {
			log.Println("Deep actions check address token error 1:", err)
			return false
		}
		return Addr.TokenLabel != ""
	}

	return false
}

func (a *Address) ScAbandonment(address string) error {

	if !crypt.IsAddressUw(address) || address == config.GenesisAddress {
		return errors.New("Deep actions smart contract abandonment error 1")
	}

	row := a.GetAddress(address)
	if row != "" {
		Addr := Address{}
		err := json.Unmarshal([]byte(row), &Addr)
		if err != nil {
			return errors.New("Deep actions smart contract abandonment error 2")
		}

		if Addr.ScKeeping {
			return errors.New("Deep actions smart contract abandonment error 3")
		}

		Addr.ScKeeping = true
		jsonString, err := json.Marshal(Addr)
		if err != nil {
			return errors.New("Deep actions smart contract abandonment error 4")
		}

		leveldb.AddressDB.Put(address, string(jsonString))
		return nil
	}

	return errors.New("Deep actions smart contract abandonment error 5")
}

func updateBalance(balance []Balance, newBalance Balance, side bool) []Balance {
	idx := -1
	for i := range balance {
		if balance[i].TokenLabel == newBalance.TokenLabel {
			idx = i
		}
	}

	newBalance.Amount, _ = apparel.Round(newBalance.Amount)

	switch side {
	case true:
		if idx >= 0 {
			balance[idx].Amount, _ = apparel.Round(balance[idx].Amount)
			balance[idx].Amount += newBalance.Amount
			balance[idx].UpdateTime = newBalance.UpdateTime
		} else {
			balance = append(balance, newBalance)
		}
		break
	case false:
		if idx >= 0 {
			if balance[idx].Amount < newBalance.Amount {
				return balance
			} else {
				balance[idx].Amount, _ = apparel.Round(balance[idx].Amount)
				balance[idx].Amount -= newBalance.Amount
				balance[idx].UpdateTime = newBalance.UpdateTime
			}
		}
		break
	}

	return balance
}
