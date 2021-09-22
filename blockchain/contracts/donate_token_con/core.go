package donate_token_con

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"node/apparel"
	"node/blockchain/contracts"
	"node/config"
	"node/crypt"
	"node/memory"
	"strconv"
)

var (
	db = contracts.Database{}

	//TxsDB = db.NewConnection("blockchain/contracts/donate_token_con/storage/donate_token_contract_txs")
	//TxDB  = db.NewConnection("blockchain/contracts/donate_token_con/storage/donate_token_contract_tx")
	//LogDB = db.NewConnection("blockchain/contracts/donate_token_con/storage/donate_token_contract_log")
	EventDB  = db.NewConnection("blockchain/contracts/donate_token_con/storage/donate_token_contract_event")
	ConfigDB = db.NewConnection("blockchain/contracts/donate_token_con/storage/donate_token_contract_config")
)

func ChangeStandard(scAddress string) error {
	scBalance := contracts.GetBalance(scAddress)
	if scBalance == nil {
		return nil
	}

	token := contracts.GetTokenInfoForScAddress(scAddress)
	if token.Id == 0 {
		return errors.New("Donate token contract error 1: token does not exist")
	}

	var txs []contracts.Tx
	for _, i := range scBalance {
		timestampD := strconv.FormatInt(apparel.TimestampUnix(), 10)

		tx := contracts.Tx{
			Type:       5,
			Nonce:      apparel.GetNonce(timestampD),
			HashTx:     "",
			Height:     config.BlockHeight,
			From:       scAddress,
			To:         token.Proposer,
			Amount:     i.Amount,
			TokenLabel: i.TokenLabel,
			Timestamp:  timestampD,
			Tax:        0,
			Signature:  nil,
			Comment: *contracts.NewComment(
				"refund_transaction",
				nil,
			),
		}
		txs = append(txs, tx)
	}

	if txs != nil && memory.IsNodeProposer() {
		for _, i := range txs {
			tx := contracts.NewTx(
				i.Type,
				i.Nonce,
				i.HashTx,
				i.Height,
				i.From,
				i.To,
				i.Amount,
				i.TokenLabel,
				i.Timestamp,
				i.Tax,
				i.Signature,
				i.Comment,
			)

			jsonString, _ := json.Marshal(contracts.Tx{
				Type:       tx.Type,
				Nonce:      tx.Nonce,
				From:       tx.From,
				To:         tx.To,
				Amount:     tx.Amount,
				TokenLabel: tx.TokenLabel,
				Comment:    tx.Comment,
			})
			tx.Signature = crypt.SignMessageWithSecretKey(config.NodeSecretKey, jsonString)

			contracts.SendTx(*tx)
			*contracts.TransactionsMemory = append(*contracts.TransactionsMemory, *tx)
		}
	}

	return nil
}