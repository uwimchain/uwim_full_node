package deep_actions

import (
	"encoding/json"
	"log"
	"node/apparel"
	"node/crypt"
	"node/storage/leveldb"
	"sort"
	"strconv"
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

func NewAddress(address string, balance []Balance, publicKey []byte, firstTxTime string, lastTxTime string) *Address {
	return &Address{
		Address:     address,
		Balance:     balance,
		PublicKey:   publicKey,
		FirstTxTime: firstTxTime,
		LastTxTime:  lastTxTime,
		TokenLabel:  "",
		ScKeeping:   false,
		Name:        "",
	}

}

func (a *Address) Create() {
	jsonString, _ := json.Marshal(a)

	leveldb.AddressDB.Put(a.Address, string(jsonString))
}

func GetAddress(addressString string) *Address {
	address := new(Address)
	addressJson := leveldb.AddressDB.Get(addressString).Value
	_ = json.Unmarshal([]byte(addressJson), address)
	return address
}

func (a *Address) Update() {
	jsonString, _ := json.Marshal(a)

	leveldb.AddressDB.Put(a.Address, string(jsonString))
}

func (a *Address) UpdateBalance(address string, balance Balance, side bool) {
	if a.Address == "" {
		publicKey, err := crypt.PublicKeyFromAddress(address)
		if err != nil {
			log.Println("Update Balance error 1:", err)
		}

		a.Address = address
		a.PublicKey = publicKey
		a.FirstTxTime = balance.UpdateTime
	}

	a.Balance = updateBalance(a.Balance, balance, side)
	a.LastTxTime = balance.UpdateTime
	a.Update()
}

func (a *Address) GetBalance() []Balance {
	return a.Balance
}

func (a *Address) CheckAddressToken() bool {
	return a.TokenLabel != ""
}

func (a *Address) GetToken() *Token {
	if a.TokenLabel == "" {
		return nil
	}

	return GetToken(a.TokenLabel)
}

func (a *Address) ScAbandonment() {
	if !a.ScKeeping {
		a.ScKeeping = true
		a.Update()
	}
}

func updateBalance(balance []Balance, newBalance Balance, side bool) []Balance {
	idx := -1
	for i := range balance {
		if balance[i].TokenLabel == newBalance.TokenLabel {
			idx = i
		}
	}

	newBalance.Amount = apparel.Round(newBalance.Amount)

	switch side {
	case true:
		if idx >= 0 {
			balance[idx].Amount = apparel.Round(balance[idx].Amount)
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
				balance[idx].Amount = apparel.Round(balance[idx].Amount)
				balance[idx].Amount -= newBalance.Amount
				balance[idx].UpdateTime = newBalance.UpdateTime
			}
		}
		break
	}
	return balance
}

func (a *Address) GetTxs() []Tx {
	if a.Address == "" {
		return nil
	}
	txsJson := leveldb.TxDB.Get(a.Address).Value
	var txs []Tx

	if txsJson == "" {
		return nil
	}

	_ = json.Unmarshal([]byte(txsJson), &txs)

	if txs == nil {
		return nil
	}

	sort.Slice(txs, func(i, j int) bool {
		timestamp1, _ := strconv.ParseInt(txs[i].Timestamp, 10, 64)
		timestamp2, _ := strconv.ParseInt(txs[j].Timestamp, 10, 64)
		return timestamp1 > timestamp2
	})

	if len(txs) > 30 {
		return txs[:30]
	} else {
		return txs
	}
}

func (a *Address) AppendTx(tx Tx) {
	txs := append(a.GetTxs(), tx)
	jsonString, _ := json.Marshal(txs)
	leveldb.TxDB.Put(a.Address, string(jsonString))
}

func (a *Address) UpdateBalanceTest(amount float64, label string, side bool) {
	if a.Address == "" {
		return
	}

	idx := -1
	for i := range a.Balance {
		if a.Balance[i].TokenLabel == label {
			idx = i
		}
	}

	amount = apparel.Round(amount)
	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)

	switch side {
	case true:
		if idx >= 0 {
			a.Balance[idx].Amount += amount
			a.Balance[idx].UpdateTime = timestamp
		} else {
			a.Balance = append(a.Balance, Balance{
				TokenLabel: label,
				Amount:     amount,
				UpdateTime: timestamp,
			})
		}
		break
	case false:
		if idx >= 0 {
			if a.Balance[idx].Amount >= amount {
				a.Balance[idx].Amount -= amount
				a.Balance[idx].UpdateTime = timestamp
			}
		}
		break
	}

	if a.FirstTxTime == "" {
		a.FirstTxTime = timestamp
	}

	a.LastTxTime = timestamp

	a.Update()
}
