package deep_actions

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"node/config"
	"node/crypt"
	"node/storage/leveldb"
	"strconv"
)

// Структура транзакции, которые записываются в блок и хранятся до этого момента в памяти ноды
// Типы транзакций:
// 1: Обычная транзакция
// 2: Награда
// 3: Действия с токенами
// 4: Создание смарт контракта
// 5: Действия смарт контракта
type Tx struct {
	Type       int64   `json:"type"`       // Тип транзакции
	Nonce      int64   `json:"nonce"`      // Уникальный идентификатор транзакции
	HashTx     string  `json:"hashTx"`     // Строка хэша данных транзакции
	Height     int64   `json:"height"`     // Высота блока, которая была действительной, когда транзакция была отправлена
	From       string  `json:"from"`       // Адрес отправителя транзакции
	To         string  `json:"to"`         // Адрес получателя транзакции
	Amount     float64 `json:"amount"`     // Количество монет, отправленных транзакцией
	TokenLabel string  `json:"tokenLabel"` // Обозначение монет, которые были отправлены
	Timestamp  string  `json:"timestamp"`  // Время отправки транзакции
	Tax        float64 `json:"tax"`        // Комиссия транзакции
	Signature  []byte  `json:"signature"`  // Подпись транзакции отправителем
	Comment    Comment `json:"comment"`    // Комментарий к транзакции
}

// Структура комментария к транзакции в нём может содержаться разная информация,
//которая зависит от типа и подтипа транзакции и валидируется в зависимости от типа транзакции и Title комментария
type Comment struct {
	Title string `json:"title"` // Заголовок комментария, в зависимости от него будут валидироваться данные в Data
	Data  []byte `json:"data"`  // Данные комментрия, хранят в себе дополнительные параметры транзакции в формате JSON
}

// Функция конструктора структуры Comment. Возвращает объект структуры Comment в зависимости от заданных параметров
func NewComment(title string, data []byte) *Comment {
	return &Comment{
		Title: title,
		Data:  data,
	}
}

// Функция конструктора структуры Tx. Возвращает объект структуры Tx в зависимости от заданных параметров
func NewTx(txType int64, nonce int64, hashTx string, height int64, from string, to string, amount float64, tokenLabel string, timestamp string, tax float64, signature []byte, comment Comment) *Tx {
	return &Tx{
		Type:       txType,
		Nonce:      nonce,
		HashTx:     hashTx,
		Height:     height,
		From:       from,
		To:         to,
		Amount:     amount,
		TokenLabel: tokenLabel,
		Timestamp:  timestamp,
		Tax:        tax,
		Signature:  signature,
		Comment:    comment,
	}
}

// Функция получения транзакций пользователя по его адресу
func (tx *Tx) GetTx(address string) string {
	return leveldb.TxDB.Get(address).Value
}

//Chain
// Структура блока
type Chain struct {
	Hash   string `json:"hash"`   // Хэш блока
	Header Header `json:"header"` // Заголовок блока
	Txs    []Tx   `json:"txs"`    // Транзакции блока
}

// Функция конструктора структуры Chain. Возвращает объект структуры Chain в зависимости от заданных параметров
func NewChain(hash string, header Header, txs []Tx) *Chain {
	return &Chain{
		Hash:   hash,
		Header: header,
		Txs:    txs,
	}
}

// Структура заголовка блока
type Header struct {
	PrevHash          string `json:"prevHash"`          // Хэш предыдущего блока
	TxCounter         int64  `json:"txCounter"`         // Количество транзакций в блоке
	Timestamp         string `json:"timestamp"`         // Время записи блока
	ProposerSignature []byte `json:"proposerSignature"` // Сигратура ноды, записавшей блок
	Proposer          string `json:"proposer"`          // Андрес ноды, записавшей блок
	Votes             []Vote `json:"votes"`             // Массив с голосами нод зазапись блока
	VoteCounter       int64  `json:"voteCounter"`       // Количество голосов в блоке
}

// Функция конструктора структуры Header. Возвращает объект структуры Header в зависимости от заданных параметров
func NewHeader(prevHash string, txCounter int64, timestamp string, proposerSignature []byte, proposer string, votes []Vote, voteCounter int64) *Header {
	return &Header{
		PrevHash:          prevHash,
		TxCounter:         txCounter,
		Timestamp:         timestamp,
		ProposerSignature: proposerSignature,
		Proposer:          proposer,
		Votes:             votes,
		VoteCounter:       voteCounter,
	}
}

// Структура голоса за запись блока
type Vote struct {
	Proposer    string `json:"proposer"`    // Адрес голосующей ноды
	Signature   []byte `json:"signature"`   // Подпись голоса голосующей нодой
	BlockHeight int64  `json:"blockHeight"` // Высота блока голосующей ноды
	Vote        bool   `json:"vote"`        // Голос ноды за запись блока
}

// Функция конструктора структуры Vote. Возвращает объект структуры Vote в зависимости от заданных параметров
func NewVote(proposer string, signature []byte, blockHeight int64, vote bool) *Vote {
	return &Vote{
		Proposer:    proposer,
		Signature:   signature,
		BlockHeight: blockHeight,
		Vote:        vote,
	}
}

// Функция записи блока в базу данных
func (c *Chain) NewChain(chain Chain) error {
	chainForJson := Chain{
		Header: Header{
			PrevHash:          chain.Header.PrevHash,
			TxCounter:         chain.Header.TxCounter,
			Timestamp:         chain.Header.Timestamp,
			ProposerSignature: chain.Header.ProposerSignature,
			Proposer:          chain.Header.Proposer,
		},
		Txs: chain.Txs,
	}

	jsonForHash, err := json.Marshal(chainForJson)
	if err != nil {
		log.Println("New Chain error: ", err)
	}

	hash := crypt.GetHash(jsonForHash)

	if hash != chain.Header.PrevHash {

		jsonString, err := json.Marshal(NewChain(hash, chain.Header, chain.Txs))
		if err != nil {
			return errors.New(fmt.Sprintf("New Chain error: %v", err))
			//log.Println("New Chain error: ", err)
		}

		leveldb.ChainDB.Put(strconv.FormatInt(config.BlockHeight, 10), string(jsonString))
	} else {
		return errors.New("New chain error: hash == prev hash")
		//log.Println("New chain error: hash == prev hash")
	}

	return nil
}

// Функция получения блока по его высоте
func (c *Chain) GetChain(height string) string {
	return leveldb.ChainDB.Get(height).Value
}

//Config
// Структура записи конфига
type Config struct {
	Key   string `json:"key"`   // Название параметра конфига
	Value string `json:"value"` // Значение параметра конфига
}

// Функция получения параметра конфига по его названию
func (c *Config) GetConfig(key string) string {
	return leveldb.ConfigDB.Get(key).Value
}

// Функция изменения параметра конфига
func (c *Config) ConfigUpdate(key string, value string) {
	leveldb.ConfigDB.Put(key, value)
}

//Apparel
// Вспомогательная функция добавления транзакции к остальным транзакциям пользователя
func AppendTx(addressTxsRow string, tx Tx) string {
	var result []byte
	var AddressTxs []Tx
	if addressTxsRow == "" {
		AddressTxs = append(AddressTxs, tx)
	} else {
		err := json.Unmarshal([]byte(addressTxsRow), &AddressTxs)
		if err != nil {
			log.Println("append Tx error:", err)
		}

		AddressTxs = append(AddressTxs, tx)
	}

	result, err := json.Marshal(AddressTxs)
	if err != nil {
		log.Println("append Tx error:", err)
	}

	return string(result)
}
