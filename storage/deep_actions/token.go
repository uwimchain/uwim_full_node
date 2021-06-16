package deep_actions

import (
	"encoding/json"
	"log"
	"node/apparel"
	"node/storage/leveldb"
	"strconv"
)

type Token struct {
	Id                  int64     `json:"tokenId"`
	Type                int64     `json:"type"`
	Label               string    `json:"label"`
	Name                string    `json:"name"`
	Proposer            string    `json:"proposer"`
	Signature           []byte    `json:"signature"`
	Emission            float64   `json:"emission"`
	Timestamp           int64     `json:"timestamp"`
	Standard            int64     `json:"standard"`
	StandardHistory     []History `json:"standard_history"`
	StandardCard        string    `json:"standard_card"`
	StandardCardHistory []History `json:"standard_card_history"`
	Card                string    `json:"card"`
	CardHistory         []History `json:"card_history"`
}

type History struct {
	Id        int64  `json:"id"`
	Timestamp string `json:"timestamp"`
	TxHash    string `json:"tx_hash"`
}

type PersonalTokenCard struct {
	FullName   string      `json:"full_name"`
	BirthDay   string      `json:"birth_day"`
	Gender     string      `json:"gender"`
	Country    string      `json:"country"`
	Region     string      `json:"region"`
	City       string      `json:"city"`
	Social     *Social     `json:"social"`
	Messengers *Messengers `json:"messengers"`
	Email      string      `json:"email"`
	Site       string      `json:"site"`
	Hashtags   string      `json:"hashtags"`
	WorkPlace  string      `json:"work_place"`
	School     string      `json:"school"`
	Education  string      `json:"education"`
	Comment    string      `json:"comment"`
}

type Social struct {
	Vk        string `json:"vk"`
	Facebook  string `json:"facebook"`
	YouTube   string `json:"you_tube"`
	Instagram string `json:"instagram"`
	Twitter   string `json:"twitter"`
}

type Messengers struct {
	WhatsUp  string `json:"whats_up"`
	Telegram string `json:"telegram"`
	Discord  string `json:"discord"`
	Snapchat string `json:"snapchat"`
	Viber    string `json:"viber"`
}

type ThxStandardCardData struct {
	FieldOfActivity string    `json:"field_of_activity"`
	Achievements    []string  `json:"achievements"`
	Portfolio       []string  `json:"portfolio"`
	Social          *Social   `json:"social"`
	Site            string    `json:"site"`
	Brand           string    `json:"brand"`
	Contacts        *Contacts `json:"contacts"`
	Description     string    `json:"description"`
}

type DropStandardCardData struct {
	FieldOfActivity string    `json:"field_of_activity"`
	Achievements    []string  `json:"achievements"`
	Portfolio       []string  `json:"portfolio"`
	Social          *Social   `json:"social"`
	Site            string    `json:"site"`
	Brand           string    `json:"brand"`
	Contacts        *Contacts `json:"contacts"`
	Conversion      float64   `json:"conversion"`
}

type StartUpStandardCardData struct {
	SubjectMatters      []string        `json:"subject_matters"`
	Team                string          `json:"team"`
	Videos              []string        `json:"videos"`
	ImplementationPlan  string          `json:"implementation_plan"`
	EventRibbon         string          `json:"event_ribbon"`
	Social              *Social         `json:"social"`
	Contacts            *Contacts       `json:"contacts"`
	ProjectName         string          `json:"project_name"`
	Comment             string          `json:"comment"`
	InvestorsConditions []string        `json:"investors_conditions"`
	Conversion          float64         `json:"conversion"`
	CollectionAmount    float64         `json:"collection_amount"`
	ListingPromises     []string        `json:"listing_promises"`
	Site                string          `json:"site"`
	AdditionalData      *AdditionalData `json:"additional_data"`
}

type BusinessStandardCardData struct {
	SubjectMatters     []string        `json:"subject_mattersssss"`
	Team               string          `json:"team"`
	Videos             []string        `json:"videos"`
	ImplementationPlan string          `json:"implementation_plan"`
	EventRibbon        string          `json:"event_ribbon"`
	Social             *Social         `json:"social"`
	Contacts           *Contacts       `json:"contacts"`
	ProjectName        string          `json:"project_name"`
	Comment            string          `json:"comment"`
	Conditions         []string        `json:"conditions"`
	Conversion         float64         `json:"conversion"`
	SalesValue         float64         `json:"sales_value"`
	ListingPromises    []string        `json:"listing_promises"`
	Site               string          `json:"site"`
	AdditionalData     *AdditionalData `json:"additional_data"`
}

type AdditionalData struct {
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

func NewToken(id int64, tType int64, label string, name string, proposer string,
	signature []byte, emission float64, timestamp int64) *Token {
	return &Token{Id: id, Type: tType, Label: label, Name: name, Proposer: proposer, Signature: signature, Emission: emission, Timestamp: timestamp}
}

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

			stringTimestamp := apparel.UnixToString(timestamp)
			A.UpdateBalance(proposer, *NewBalance(label, emission, stringTimestamp), true)
		}
	}
}

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

func (t *Token) GetToken(tokenLabel string) string {
	return leveldb.TokenDb.Get(tokenLabel).Value
}

func (t *Token) CheckToken(tokenLabel string) bool {
	return leveldb.TokenDb.Has(tokenLabel)
}

func (t *Token) GetTokens() []Token {
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

func (t *Token) AutoIncrement() int64 {
	lastId := leveldb.ConfigDB.Get("token_id").Value
	if lastId != "" {
		result := apparel.ParseInt64(lastId)
		return result + 1
	} else {
		return 1
	}
}
