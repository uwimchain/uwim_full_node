package deep_actions

import (
	"encoding/json"
	"log"
	"node/apparel"
	"node/config"
	"node/crypt"
	"node/storage/leveldb"
	"strconv"
)

type String string

type Token struct {
	Id int64 `json:"tokenId"`
	// 0 - Personal
	// 1 - Team
	// 2 - Contract
	Type      int64   `json:"type"`
	Label     string  `json:"label"`
	Name      string  `json:"name"`
	Proposer  string  `json:"proposer"`
	Signature []byte  `json:"signature"`
	Emission  float64 `json:"emission"`
	Timestamp String  `json:"timestamp"`
	// 0 - My
	// 1 - Donate
	// 3 - StartUp
	// 4 - Business
	// 5 - Trade
	// 6 - Payable
	// 7 - Nft
	Standard            int64     `json:"standard"`
	StandardHistory     []History `json:"standard_history"`
	StandardCard        string    `json:"standard_card"`
	StandardCardHistory []History `json:"standard_card_history"`
	Card                string    `json:"card"`
	CardHistory         []History `json:"card_history"`
}

type Tokens []Token

type History struct {
	Id        int64  `json:"id"`
	Timestamp string `json:"timestamp"`
	TxHash    string `json:"tx_hash"`
}

type PersonalTokenCard struct {
	FullName   string      `json:"full_name"`
	BirthDay   string      `json:"birth_day"`
	Gender     String      `json:"gender"`
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

type DonateStandardCardData struct {
	FieldOfActivity string    `json:"field_of_activity"`
	Achievements    []string  `json:"achievements"`
	Portfolio       []string  `json:"portfolio"`
	Social          *Social   `json:"social"`
	Site            string    `json:"site"`
	Brand           string    `json:"brand"`
	Contacts        *Contacts `json:"contacts"`
}

type BuyTokenSign struct {
	NodeAddress string `json:"node_address"`
}

func NewBuyTokenSign(nodeAddress string) *BuyTokenSign {
	return &BuyTokenSign{NodeAddress: nodeAddress}
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
	ListingPromises     []string        `json:"listing_promises"`
	Site                string          `json:"site"`
	AdditionalData      *AdditionalData `json:"additional_data"`
}

type BusinessStandardCardData struct {
	Team               string          `json:"team"`
	Videos             []string        `json:"videos"`
	ImplementationPlan string          `json:"implementation_plan"`
	EventRibbon        string          `json:"event_ribbon"`
	Social             *Social         `json:"social"`
	Contacts           *Contacts       `json:"contacts"`
	ProjectName        string          `json:"project_name"`
	Comment            string          `json:"comment"`
	Site               string          `json:"site"`
	AdditionalData     *AdditionalData `json:"additional_data"`
	Conditions         []string        `json:"conditions"`
	SubjectMatters     []string        `json:"subject_matters"`
	ListingPromises    []string        `json:"listing_promises"`
}

type TradeStandardCardData struct {
	Team               string          `json:"team"`
	Videos             []string        `json:"videos"`
	ImplementationPlan string          `json:"implementation_plan"`
	EventRibbon        string          `json:"event_ribbon"`
	Social             *Social         `json:"social"`
	Contacts           *Contacts       `json:"contacts"`
	ProjectName        string          `json:"project_name"`
	Comment            string          `json:"comment"`
	Site               string          `json:"site"`
	AdditionalData     *AdditionalData `json:"additional_data"`
	Conditions         []string        `json:"conditions"`
	SubjectMatters     []string        `json:"subject_matters"`
	ListingPromises    []string        `json:"listing_promises"`
}

type NftStandardCardData struct {
	Team               string          `json:"team"`
	Videos             []string        `json:"videos"`
	ImplementationPlan string          `json:"implementation_plan"`
	EventRibbon        string          `json:"event_ribbon"`
	Social             *Social         `json:"social"`
	Contacts           *Contacts       `json:"contacts"`
	ProjectName        string          `json:"project_name"`
	Comment            string          `json:"comment"`
	Site               string          `json:"site"`
	AdditionalData     *AdditionalData `json:"additional_data"`
	Conditions         []string        `json:"conditions"`
	SubjectMatters     []string        `json:"subject_matters"`
}

type Partner struct {
	Address string  `json:"address"`
	Percent float64 `json:"percent"`
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

func NewToken(tType int64, label string, name string, proposer string,
	signature []byte, emission float64, timestamp string, standard int64) *Token {
	return &Token{
		Type:      tType,
		Label:     label,
		Name:      name,
		Proposer:  proposer,
		Signature: signature,
		Emission:  emission,
		Timestamp: String(timestamp),
		Standard:  standard,
	}
}

func (t *Token) Create() {
	t.Id = autoincrement()
	jsonString, err := json.Marshal(t)
	if err != nil {
		log.Println("Deep actions new token error 2: ", err)
	}

	if t.Type == 2 {
		t.Emission = 0
	}

	leveldb.TokenDb.Put(t.Label, string(jsonString))
	leveldb.TokenIdsDb.Put(strconv.FormatInt(t.Id, 10), t.Label)

	address := GetAddress(t.Proposer)
	address.TokenLabel = t.Label
	address.Update()

	leveldb.ConfigDB.Put("token_id", strconv.FormatInt(t.Id, 10))

	timestamp := strconv.FormatInt(apparel.TimestampUnix(), 10)
	address.UpdateBalance(t.Proposer, t.Emission, t.Label, timestamp, true)
}

func (t *Token) SetSignature(secretKey []byte) {
	jsonString, _ := json.Marshal(t)

	t.Signature = crypt.SignMessageWithSecretKey(secretKey, jsonString)
}

func (t *Token) Update() {
	jsonString, _ := json.Marshal(t)
	leveldb.TokenDb.Put(t.Label, string(jsonString))
}

func (t *Token) RenameToken(newName string) {
	t.Name = newName
	t.Update()
}

func (t *Token) ChangeTokenStandard(newStandard int64, timestamp string, txHash string) {
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

	t.Update()
}

func (t *Token) AddTokenEmission(addEmissionAmount float64) {
	if t.Emission+addEmissionAmount > config.MaxEmission {
		return
	}

	t.Emission += addEmissionAmount

	t.Update()
}

func (t *Token) FillTokenCard(newCardData []byte, timestamp string, txHash string) {
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

	t.Update()
}

func (t *Token) FillTokenStandardCard(newStandardCardData []byte, timestamp string, txHash string) {
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

	t.Update()
}

func (t *Token) GetStandardCard() map[string]interface{} {
	standardCardData := make(map[string]interface{})
	_ = json.Unmarshal([]byte(t.StandardCard), &standardCardData)

	return standardCardData
}

func GetToken(label string) *Token {
	tokenJson := leveldb.TokenDb.Get(label).Value
	token := new(Token)
	_ = json.Unmarshal([]byte(tokenJson), token)

	return token
}

func (s *String) UnmarshalJSON(b []byte) error {
	var item interface{}
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}

	switch v := item.(type) {
	case int64:
		*s = String(strconv.FormatInt(v, 10))
		break
	case float64:
		*s = String(strconv.Itoa(int(v)))
		break
	case string:
		*s = String(v)
	}
	return nil
}

func CheckToken(tokenLabel string) bool {
	return leveldb.TokenDb.Has(tokenLabel)
}

func GetAllTokens() Tokens {
	rows := leveldb.TokenDb.GetAll("")

	var tokens Tokens

	if rows != nil {
		for _, row := range rows {
			token := Token{}
			_ = json.Unmarshal([]byte(row.Value), &token)

			tokens = append(tokens, token)
		}
	}

	return tokens
}

func autoincrement() int64 {
	lastId := leveldb.ConfigDB.Get("token_id").Value
	if lastId != "" {
		result, _ := strconv.ParseInt(lastId, 10, 64)
		return result + 1
	} else {
		return 1
	}
}
