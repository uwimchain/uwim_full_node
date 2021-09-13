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

// Структура данных аккаунта пользователя в системе
type Address struct {
	Address     string    `json:"address"`     // Адрес пользователя
	Balance     []Balance `json:"balance"`     // Массив с балансом, содержит в себе информацию о количестве разных валют пользователя
	PublicKey   []byte    `json:"publicKey"`   // Публичный ключ пользователя
	FirstTxTime string    `json:"firstTxTime"` // Время первой транзакции, транзакция с перечислением монет на данный аккаунт
	LastTxTime  string    `json:"lastTxTime"`  // Время крайней транзакции вне зависимости от того, отправили монеты пользователю или он сам отправил транзакцию
	TokenLabel  string    `json:"tokenLabel"`  // Обозначения токена пользователя, пока токен не создан остаётся пустым
	ScKeeping   bool      `json:"sc_keeping"`  // Наличие смарт-контракта у пользователя
	Name        string    `json:"name"`        // Ник аккаунта пользователя
}

// Структура баланса одного токена пользователя
type Balance struct {
	TokenLabel string  `json:"tokenLabel"` // Обозначение токена
	Amount     float64 `json:"amount"`     // Количество монет
	UpdateTime string  `json:"updateTime"` // Время последнего обновления баланса этого токена
}

// Конструктор структуры Balance. Возвращает объект структуры Balance с задаными данными
func NewBalance(tokenLabel string, amount float64, updateTime string) *Balance {
	return &Balance{TokenLabel: tokenLabel, Amount: amount, UpdateTime: updateTime}
}

// Функция создания нового адреса в базе данных
func (a *Address) NewAddress(address string, balance []Balance, publicKey []byte, firstTxTime string,
	lastTxTime string) {
	//jsonString, err := json.Marshal(NewAddress(address, balance, publicKey, firstTxTime, lastTxTime, ""))
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

// Функция получения данных аккаунта по адресу из базы данных
func (a *Address) GetAddress(address string) string {
	return leveldb.AddressDB.Get(address).Value
}

// Функция изменения баланса пользователя
func (a *Address) UpdateBalance(address string, balance Balance, side bool) {
	// Получение данных пользователя по адресу из базы данных
	row := a.GetAddress(address)

	// Если пользователя нет в системе, то создаётся аккаунт
	if row == "" {
		publicKey, err := crypt.PublicKeyFromAddress(address)
		if err != nil {
			log.Println("Update Balance error 1:", err)
		}

		a.NewAddress(address, nil, publicKey, balance.UpdateTime, balance.UpdateTime)
	}

	// Повторное получение данных пользователя из базы данных по адресу
	row = a.GetAddress(address)
	Addr := Address{}
	err := json.Unmarshal([]byte(row), &Addr)
	if err != nil {
		log.Println("Update Balance error 2:", err)
	}

	// Изменение баланса пользователя в зависимости от введённых данных
	Addr.Balance = updateBalance(Addr.Balance, balance, side)

	// Выставление даты крайнего изменения баланса выбранной монеты
	Addr.LastTxTime = balance.UpdateTime

	jsonString, err := json.Marshal(Addr)
	if err != nil {
		log.Println("Update Balance error 3:", err)
	}

	// Помещение обновлённого баланса пользователя в базу данных
	leveldb.AddressDB.Put(address, string(jsonString))
}

// Функция проверки наличия у пользователя собственного токена по адресу
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

// Функция для отказа от смарт-контракта
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

//Apparel
// Вспомогательная функция для получения номера токена в балансе пользователя
/*func findTokenInBalance(balance []Balance, token string) int {
	for i := range balance {
		if balance[i].TokenLabel == token {
			return i
		}
	}

	return len(balance)
}*/

// Вспомогательная функция изменения баланса в зависимости от заданых параметров
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

	/*	if idx := findTokenInBalance(balance, newBalance.TokenLabel); idx != len(balance) {
		if side {
			balance[idx].Amount += newBalance.Amount
			balance[idx].UpdateTime = newBalance.UpdateTime
		} else {
			if balance[idx].Amount < newBalance.Amount {
				return balance
			} else {
				balance[idx].Amount -= newBalance.Amount
				balance[idx].UpdateTime = newBalance.UpdateTime
			}
		}
		} else {
			balance = append(balance, newBalance)
		}*/

	return balance
}
