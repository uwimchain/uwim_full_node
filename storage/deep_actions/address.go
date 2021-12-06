package deep_actions

import (
	"encoding/json"
	"log"
	"node/apparel"
	"node/config"
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

func (a *Address) UpdateBalance(address string, amount float64, tokenLabel, timestamp string, side bool) {
	if a.Address == "" {
		publicKey, err := crypt.PublicKeyFromAddress(address)
		if err != nil {
			log.Println("Update Balance error 1:", err)
		}

		a.Address = address
		a.PublicKey = publicKey
		a.FirstTxTime = timestamp
	}

	a.Balance = updateBalance(a.Balance, Balance{
		TokenLabel: tokenLabel,
		Amount:     amount,
		UpdateTime: timestamp,
	}, side)
	a.LastTxTime = timestamp

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

func (a *Address) GetTxs() Txs {
	if a.Address == "" {
		return nil
	}
	txsJson := leveldb.TxDB.Get(a.Address).Value
	var txs Txs

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

	if len(txs) > config.BalanceTransactionsLimit {
		return txs[:config.BalanceTransactionsLimit]
	} else {
		return txs
	}
}

func (a *Address) GetAllTxs() Txs {
	if a.Address == "" {
		return nil
	}
	txsJson := leveldb.TxDB.Get(a.Address).Value
	var txs Txs

	if txsJson == "" {
		return nil
	}

	_ = json.Unmarshal([]byte(txsJson), &txs)

	if txs == nil {
		return nil
	}

	return txs
}

func (a *Address) AppendTx(tx Tx) {
	txs := a.GetAllTxs()
	txs = append(txs, tx)
	jsonString, _ := json.Marshal(txs)
	leveldb.TxDB.Put(a.Address, string(jsonString))
}
