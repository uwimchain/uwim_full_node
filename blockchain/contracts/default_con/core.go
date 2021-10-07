package default_con

import (
	"encoding/base64"
	"encoding/json"
	"node/blockchain/contracts"
	"strconv"
)

var (
	db = contracts.Database{}

	EventDB         = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_event")
	ConfigDB        = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_config")
	TokenDB         = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_token")
	LabelTokenElsDB = db.NewConnection("blockchain/contracts/default_con/storage/default_contract_label_token_els")
)

type NftTokenEl struct {
	Id          int64   `json:"id"` // unique autoincrement
	ParentLabel string  `json:"parent_label"`
	Name        string  `json:"name"`
	Owner       string  `json:"owner"` // uw address
	Price       float64 `json:"price"`
	Data        string  `json:"data"` // base64
	Timestamp   string  `json:"timestamp"`
}

type NftTokenEls []NftTokenEl

func NewNftTokenEl(name string, owner string, price float64, data string, timestamp string) *NftTokenEl {

	return &NftTokenEl{Name: name, Owner: owner, Price: price, Data: data, Timestamp: timestamp}
}

func GetNftTokenElForId(id int64) *NftTokenEl {
	tokenElJson := TokenDB.Get(strconv.FormatInt(id, 10)).Value
	tokenEl := NftTokenEl{}
	_ = json.Unmarshal([]byte(tokenElJson), &tokenEl)

	return &tokenEl
}

func GetNftAllTokensEls() NftTokenEls {
	tokensElsJson := TokenDB.GetAll("")
	if tokensElsJson == nil {
		return nil
	}
	tokensEls := NftTokenEls{}

	for _, i := range tokensElsJson {
		tokenEl := NftTokenEl{}
		_ = json.Unmarshal([]byte(i.Value), &tokenEl)
		tokensEls = append(tokensEls, tokenEl)
	}

	return tokensEls
}

func GetNftTokenElsForAddress(address string) NftTokenEls {
	addressObj := contracts.GetAddress(address)
	if addressObj.TokenLabel == "" {
		return nil
	}

	return GetNftTokenElsForParentLabel(addressObj.TokenLabel)
}

func GetNftTokenElsForParentLabel(label string) NftTokenEls {
	tokenElsIdsJson := LabelTokenElsDB.Get(label).Value
	if tokenElsIdsJson == "" {
		return nil
	}
	var tokenElsIds []int64
	_ = json.Unmarshal([]byte(tokenElsIdsJson), &tokenElsIds)

	if tokenElsIds == nil {
		return nil
	}

	tokenEls := NftTokenEls{}
	for _, i := range tokenElsIds {
		tokenEl := GetNftTokenElForId(i)
		if tokenEl == nil {
			continue
		}

		tokenEls = append(tokenEls, *tokenEl)
	}

	return tokenEls
}

func (tEl *NftTokenEl) Create() {
	// set autoincrement unique Id
	tEl.setAutoincrement()

	// set parent label
	tEl.setParentLabel()

	// encode nft token element data to base64
	tEl.Data = base64.StdEncoding.EncodeToString([]byte(tEl.Data))

	jsonString, _ := json.Marshal(tEl)
	TokenDB.Put(strconv.FormatInt(tEl.Id, 10), string(jsonString))

	tEl.addElToLabelElsList()
}

func (tEl *NftTokenEl) Update() {
	jsonString, _ := json.Marshal(tEl)
	TokenDB.Put(strconv.FormatInt(tEl.Id, 10), string(jsonString))
}

func (tEl *NftTokenEl) addElToLabelElsList() {
	labelTokenElsJson := LabelTokenElsDB.Get(tEl.ParentLabel).Value
	var labelTokenEls []int64
	_ = json.Unmarshal([]byte(labelTokenElsJson), &labelTokenEls)

	labelTokenEls = append(labelTokenEls, tEl.Id)
	jsonLabelTokenEls, _ := json.Marshal(labelTokenEls)
	LabelTokenElsDB.Put(tEl.ParentLabel, string(jsonLabelTokenEls))
}

func (tEl *NftTokenEl) setAutoincrement() {
	tokensElsJson := TokenDB.GetAll("")
	tEl.Id = int64(len(tokensElsJson) + 1)
}

func (tEl *NftTokenEl) setParentLabel() {
	address := contracts.GetAddress(tEl.Owner)
	parentToken := contracts.GetToken(address.TokenLabel)
	tEl.ParentLabel = parentToken.Label
}

func getParentTokensElsCount(label string) int {
	parentTokenElsIdsJson := LabelTokenElsDB.Get(label).Value
	if parentTokenElsIdsJson == "" {
		return 0
	}

	var parentTokenElsIds []int64
	_ = json.Unmarshal([]byte(parentTokenElsIdsJson), &parentTokenElsIds)

	return len(parentTokenElsIds)
}
