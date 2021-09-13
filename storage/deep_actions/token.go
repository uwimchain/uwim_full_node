package deep_actions

import (
	"encoding/json"
	"log"
	"node/apparel"
	"node/storage/leveldb"
	"strconv"
)

var (
	A Address
)

// Структура токена
type Token struct {
	Id int64 `json:"tokenId"` // Уникальный идентификатор токена в базе данных
	// 0 - Personal
	// 1 - Team
	// 2 - Nft
	Type      int64   `json:"type"`      // Тип токена
	Label     string  `json:"label"`     // Обозначение токена
	Name      string  `json:"name"`      // Наименование токена
	Proposer  string  `json:"proposer"`  // Создатель токена
	Signature []byte  `json:"signature"` // Подпись токена создателем
	Emission  float64 `json:"emission"`  // Эмиссия токена
	Timestamp int64   `json:"timestamp"` // Время записи токена в базу данных
	// 0 - My
	// 1 - Donate
	// 3 - StartUp
	// 4 - Business
	// 5 - Trade
	// 6 - Payable
	Standard            int64     `json:"standard"`              // Стандарт токена
	StandardHistory     []History `json:"standard_history"`      // История изменений стандартов токена
	StandardCard        string    `json:"standard_card"`         // Карточка стандарта
	StandardCardHistory []History `json:"standard_card_history"` // История изменений карточки стандарта токена
	Card                string    `json:"card"`                  // Карточка токена
	CardHistory         []History `json:"card_history"`          // История изменений карточки токена
}

type History struct {
	Id        int64  `json:"id"`
	Timestamp string `json:"timestamp"`
	TxHash    string `json:"tx_hash"`
}

type PersonalTokenCard struct {
	FullName   string      `json:"full_name"`  // Фамилия имя Отчество
	BirthDay   string      `json:"birth_day"`  // Дата рождения
	Gender     string      `json:"gender"`     // Пол
	Country    string      `json:"country"`    // Страна
	Region     string      `json:"region"`     // Область
	City       string      `json:"city"`       // Город
	Social     *Social     `json:"social"`     // Социальные сети
	Messengers *Messengers `json:"messengers"` // Месседжеры
	Email      string      `json:"email"`      // Электронная почта
	Site       string      `json:"site"`       // Сайт
	Hashtags   string      `json:"hashtags"`   // Интересы (хэштэги)
	WorkPlace  string      `json:"work_place"` // Место работы
	School     string      `json:"school"`     // Школа
	Education  string      `json:"education"`  // Образование
	Comment    string      `json:"comment"`    // Комментарий
}

type Social struct {
	Vk        string `json:"vk"`        //
	Facebook  string `json:"facebook"`  //
	YouTube   string `json:"you_tube"`  //
	Instagram string `json:"instagram"` //
	Twitter   string `json:"twitter"`   //
}

type Messengers struct {
	WhatsUp  string `json:"whats_up"` //
	Telegram string `json:"telegram"` //
	Discord  string `json:"discord"`  //
	Snapchat string `json:"snapchat"` //
	Viber    string `json:"viber"`    //
}

type DonateStandardCardData struct {
	FieldOfActivity string    `json:"field_of_activity"` //
	Achievements    []string  `json:"achievements"`      //
	Portfolio       []string  `json:"portfolio"`         // Портфолио
	Social          *Social   `json:"social"`            // Социальные сети
	Site            string    `json:"site"`              // Ссылка на сайт
	Brand           string    `json:"brand"`             // Название брэнда
	Contacts        *Contacts `json:"contacts"`          // Контакты
	Conversion      float64   `json:"conversion"`        // Курс покупки токена
	MaxBuy          float64   `json:"max_buy"`           // Максимальный объём покупки донатных токенов
	//Titles          []string  `json:"titles"`            // Список, на что собирается донат
}

type BuyTokenSign struct {
	NodeAddress string `json:"node_address"`
}

func NewBuyTokenSign(nodeAddress string) *BuyTokenSign {
	return &BuyTokenSign{NodeAddress: nodeAddress}
}

type StartUpStandardCardData struct {
	SubjectMatters      []string        `json:"subject_matters"`      //
	Team                string          `json:"team"`                 //
	Videos              []string        `json:"videos"`               //
	ImplementationPlan  string          `json:"implementation_plan"`  //
	EventRibbon         string          `json:"event_ribbon"`         //
	Social              *Social         `json:"social"`               //
	Contacts            *Contacts       `json:"contacts"`             //
	ProjectName         string          `json:"project_name"`         //
	Comment             string          `json:"comment"`              //
	InvestorsConditions []string        `json:"investors_conditions"` //
	Conversion          float64         `json:"conversion"`           //
	CollectionAmount    float64         `json:"collection_amount"`    //
	ListingPromises     []string        `json:"listing_promises"`     //
	Site                string          `json:"site"`                 //
	AdditionalData      *AdditionalData `json:"additional_data"`      //
}

type BusinessStandardCardData struct {
	Team               string    `json:"team"`                //
	Videos             []string  `json:"videos"`              //
	ImplementationPlan string    `json:"implementation_plan"` //
	EventRibbon        string    `json:"event_ribbon"`        //
	Social             *Social   `json:"social"`              //
	Contacts           *Contacts `json:"contacts"`            //
	ProjectName        string    `json:"project_name"`        //
	Comment            string    `json:"comment"`             //
	Site               string    `json:"site"`                //

	// token data for sale
	AdditionalData *AdditionalData `json:"additional_data"` //

	// select lists
	Conditions      []string `json:"conditions"`       //
	SubjectMatters  []string `json:"subject_matters"`  //
	ListingPromises []string `json:"listing_promises"` //

	// sale
	Conversion float64 `json:"conversion"`  // token sale course
	SalesValue float64 `json:"sales_value"` // count tokens for sale

	Changes bool `json:"changes"` // Change card data

	Partners []Partner `json:"partners"` // Partners data
}

type Partner struct {
	Address string  `json:"address"` // partner uwim address
	Percent float64 `json:"percent"` // partner percent
}

type AdditionalData struct {
	// 5: trade
	// 6: payable
	Type int64                     `json:"type"`
	Data *TradePayableStandardData `json:"data"`
}

type TradePayableStandardData struct {
	EmissionTerms                     []string `json:"emission_terms"`
	ListingTokensRate                 float64  `json:"listing_tokens_rate"`
	PreIssuePossibility               bool     `json:"pre_issue_possibility"`
	LiquidityOrganizingPoolConditions []string `json:"liquidity_organizing_pool_conditions"`
}

type Contacts struct {
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

// Конструктор структуры Token. Возвращает объект структуры Token с задаными данными
func NewToken(id int64, tType int64, label string, name string, proposer string,
	signature []byte, emission float64, timestamp int64) *Token {
	return &Token{Id: id, Type: tType, Label: label, Name: name, Proposer: proposer, Signature: signature, Emission: emission, Timestamp: timestamp}
}

// Функция добавления нового токена в базу данных
func (t *Token) NewToken(tType int64, label string, name string, proposer string,
	signature []byte, emission float64, timestamp int64) {
	if t.CheckToken(label) {
		log.Println("deep actions new token error 1: token with this label is exists in database")
	} else {
		if id := t.AutoIncrement(); id != 0 {
			jsonString, err := json.Marshal(NewToken(id, tType, label, name, proposer, signature, emission, timestamp))
			if err != nil {
				log.Println("Deep actions new token error 2: ", err)
			}

			leveldb.TokenDb.Put(label, string(jsonString))
			leveldb.TokenIdsDb.Put(strconv.FormatInt(id, 10), label)

			Addr := Address{}
			err = json.Unmarshal([]byte(A.GetAddress(proposer)), &Addr)
			if err != nil {
				log.Println("Deep actions new token error 3:", err)
			}
			if Addr.TokenLabel == "" {
				Addr.TokenLabel = label
				jsonString, err = json.Marshal(Addr)
				if err != nil {
					log.Println("Deep actions new token error 4:", err)
				}
				leveldb.AddressDB.Put(proposer, string(jsonString))
			}

			leveldb.ConfigDB.Put("token_id", strconv.FormatInt(id, 10))

			//Пополнение баланса создателя токена после его добавления в базу данных на заданную эмиссию
			timestampD := strconv.FormatInt(timestamp, 10)
			A.UpdateBalance(proposer, *NewBalance(label, emission, timestampD), true)
		}
	}
}

// Функция переименования токена
func (t *Token) RenameToken(label string, newName string) {
	row := t.GetToken(label)
	if row == "" {
		log.Println("Deep actions rename token error 1: token with this label does not exists in database")
	} else {
		err := json.Unmarshal([]byte(row), &t)
		if err != nil {
			log.Println("Deep actions rename token error 2:", err)
		} else {
			t.Name = newName
			jsonString, err := json.Marshal(t)
			if err != nil {
				log.Println("Deep actions rename token error 3:", err)
			} else {
				leveldb.TokenDb.Put(label, string(jsonString))
			}
		}
	}
}

// Функция для изменения стандарта токена
func (t *Token) ChangeTokenStandard(label string, newStandard int64, timestamp string, txHash string) {
	row := t.GetToken(label)
	if row == "" {
		log.Println("Deep actions change token standard error 1: token with this label does not exists in database")
	} else {
		err := json.Unmarshal([]byte(row), &t)
		if err != nil {
			log.Println("Deep actions change token standard error 2:", err)
		} else {
			t.Standard = newStandard
			if t.StandardHistory != nil {
				t.StandardHistory = append(t.StandardHistory, History{
					Id:        t.StandardHistory[len(t.StandardHistory)-1].Id + 1,
					Timestamp: timestamp,
					TxHash:    txHash,
				})
			} else {
				t.StandardHistory = append(t.StandardHistory, History{
					Id:        1,
					Timestamp: timestamp,
					TxHash:    txHash,
				})
			}

			jsonString, err := json.Marshal(t)
			if err != nil {
				log.Println("Deep actions change token standard error 3:", err)
			} else {
				leveldb.TokenDb.Put(label, string(jsonString))
			}
		}
	}
}

// Функция для заполнения карточки токена
func (t *Token) FillTokenCard(label string, newCardData []byte, timestamp string, txHash string) {
	row := t.GetToken(label)
	if row == "" {
		log.Println("Deep actions fill token card error 1: token with this label does not exists in database")
	} else {
		err := json.Unmarshal([]byte(row), &t)
		if err != nil {
			log.Println("Deep actions fill token card error 2:", err)
		} else {
			t.Card = string(newCardData)
			if t.CardHistory != nil {
				t.CardHistory = append(t.CardHistory, History{
					Id:        t.CardHistory[len(t.CardHistory)-1].Id + 1,
					Timestamp: timestamp,
					TxHash:    txHash,
				})
			} else {
				t.CardHistory = append(t.CardHistory, History{
					Id:        1,
					Timestamp: timestamp,
					TxHash:    txHash,
				})
			}

			jsonString, err := json.Marshal(t)
			if err != nil {
				log.Println("Deep actions fill token card error 3:", err)
			} else {
				leveldb.TokenDb.Put(label, string(jsonString))
			}
		}
	}
}

// Функция для заполнения карточки стандарта токена
func (t *Token) FillTokenStandardCard(label string, newStandardCardData []byte, timestamp string, txHash string) {
	row := t.GetToken(label)
	if row == "" {
		log.Println("Deep actions fill token standard card error 1: token with this label does not exists in database")
	} else {
		err := json.Unmarshal([]byte(row), &t)
		if err != nil {
			log.Println("Deep actions fill token standard card error 2:", err)
		} else {
			t.StandardCard = string(newStandardCardData)
			if t.StandardCardHistory != nil {
				t.StandardCardHistory = append(t.StandardCardHistory, History{
					Id:        t.StandardCardHistory[len(t.StandardCardHistory)-1].Id + 1,
					Timestamp: timestamp,
					TxHash:    txHash,
				})
			} else {
				t.StandardCardHistory = append(t.StandardCardHistory, History{
					Id:        1,
					Timestamp: timestamp,
					TxHash:    txHash,
				})
			}

			jsonString, err := json.Marshal(t)
			if err != nil {
				log.Println("Deep actions fill token standard card error 3:", err)
			} else {
				leveldb.TokenDb.Put(label, string(jsonString))
			}
		}
	}
}

// Функция получения данных токена по его обозначению
func (t *Token) GetToken(tokenLabel string) string {
	return leveldb.TokenDb.Get(tokenLabel).Value
}

// Функция проверки наличия токена в базе данных
func (t *Token) CheckToken(tokenLabel string) bool {
	return leveldb.TokenDb.Has(tokenLabel)
}

// Функция получения полного списка токенов из базы данных
func (t *Token) GetAllTokens() []Token {
	rows := leveldb.TokenDb.GetAll("")

	var tokens []Token

	if rows != nil {
		for _, row := range rows {
			token := Token{}

			err := json.Unmarshal([]byte(row.Value), &token)
			if err != nil {
				log.Println("Get Tokens error: ", err)
			}

			tokens = append(tokens, token)
		}
	}

	return tokens
}

// Функция для получения значения id для нового токена
func (t *Token) AutoIncrement() int64 {
	lastId := leveldb.ConfigDB.Get("token_id").Value
	if lastId != "" {
		result := apparel.ParseInt64(lastId)
		return result + 1
	} else {
		return 1
	}
}