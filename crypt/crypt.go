package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"golang.org/x/crypto/pbkdf2"
	"node/metrics"
	"strings"
)

func IsAddressSmartContract(address string) bool {
	if len(address) != 61 {
		return false
	}

	return strings.HasPrefix(address, metrics.SmartContractPrefix)
}

func IsAddressUw(address string) bool {
	if len(address) != 61 {
		return false
	}

	return strings.HasPrefix(address, metrics.AddressPrefix)
}

func IsAddressNode(address string) bool {
	if len(address) != 61 {
		return false
	}

	return strings.HasPrefix(address, metrics.NodePrefix)
}

//Address
func AddressFromPublicKey(hrp string, publicKey []byte) string {
	address, _ := bech32Encode(hrp, publicKey)
	return address
}

func AddressFromAnotherAddress(hrp string, anotherAddress string) string {
	publicKey, _ := PublicKeyFromAddress(anotherAddress)
	return AddressFromPublicKey(hrp, publicKey)
}

func NodeAddressFromMnemonic(mnemonic string) string {
	return AddressFromPublicKey(metrics.NodePrefix, PublicKeyFromSecretKey(SecretKeyFromSeed(SeedFromMnemonic(mnemonic))))
}

func AddressFromMnemonic(mnemonic string) string {
	return AddressFromPublicKey(metrics.AddressPrefix, PublicKeyFromSecretKey(SecretKeyFromSeed(SeedFromMnemonic(mnemonic))))
}

func ScAddressFromMnemonic(mnemonic string) string {
	return AddressFromPublicKey(metrics.SmartContractPrefix, PublicKeyFromSecretKey(SecretKeyFromSeed(SeedFromMnemonic(mnemonic))))
}

//Signature
func SignMessageWithSecretKey(secretKey []byte, message []byte) []byte {
	return ed25519.Sign(secretKey, message)
}

func VerifySign(publicKey []byte, data []byte, signature []byte) bool {
	return ed25519.Verify(publicKey, data, signature)
}

//Hash
func GetHash(jsonString []byte) string {
	alg := sha256.New()
	alg.Write(jsonString)
	return hex.EncodeToString(alg.Sum(nil))
}

//Seed
func SeedFromMnemonic(mnemonic string) []byte {
	return pbkdf2.Key([]byte(mnemonic), nil, 2048, 32, sha512.New)
}

//SecretKey
func SecretKeyFromSeed(seed []byte) []byte {
	return ed25519.NewKeyFromSeed(seed)
}

//PublicKey
func PublicKeyFromSecretKey(secretKey []byte) []byte {
	publicKey := make([]byte, ed25519.PublicKeySize)
	copy(publicKey, secretKey[32:])

	return publicKey
}

// TransactionRaw

type TxRaw struct {
	Nonce      int64        `json:"nonce"`
	From       string       `json:"from"`
	To         string       `json:"to"`
	Amount     float64      `json:"amount"`
	TokenLabel string       `json:"tokenLabel"`
	Type       int64        `json:"type"`
	Signature  string       `json:"signature"`
	Comment    TxRawComment `json:"comment"`
}

type TxRawComment struct {
	Title string `json:"title"`
	Data  string `json:"data"`
}

var TransactionRawKey = []byte{139, 111, 224, 92, 142, 122, 138, 224, 138, 118, 30, 229, 209, 155, 193, 186, 180, 234, 69, 249, 75, 71, 195, 105, 20, 61, 211, 13, 104, 253, 72, 5}
var TransactionRawIv = []byte{22, 129, 2, 139, 42, 15, 11, 131, 158, 197, 170, 43, 114, 14, 178, 167}

func DecodeTransactionRaw(transactionRaw string) (*TxRaw, error) {
	cipherText, err := base64.StdEncoding.DecodeString(transactionRaw)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("erorr 1: %v", err))
	}

	cipherBlock, err := aes.NewCipher(TransactionRawKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("erorr 2: %v", err))
	}

	if len(cipherText)%aes.BlockSize != 0 {
		return nil, errors.New(fmt.Sprintf("Cipher text (len=%d) is not a multiple of the block size (%d)", len(cipherText), aes.BlockSize))
	}

	mode := cipher.NewCBCDecrypter(cipherBlock, TransactionRawIv)
	mode.CryptBlocks(cipherText, cipherText)

	txRaw := TxRaw{}

	stringCipherText := strings.TrimSpace(string(cipherText))

	stringCipherText = strings.ReplaceAll(stringCipherText, "\x01", "")
	stringCipherText = strings.ReplaceAll(stringCipherText, "\x02", "")
	stringCipherText = strings.ReplaceAll(stringCipherText, "\x03", "")
	stringCipherText = strings.ReplaceAll(stringCipherText, "\x04", "")
	stringCipherText = strings.ReplaceAll(stringCipherText, "\x10", "")
	stringCipherText = strings.ReplaceAll(stringCipherText, "\x0f", "")
	stringCipherText = strings.ReplaceAll(stringCipherText, "\x0e", "")

	err = json.Unmarshal([]byte(stringCipherText), &txRaw)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("erorr 3: %v", err))
	}
	return &txRaw, nil
}

func PublicKeyFromAddress(address string) ([]byte, error) {
	_, publicKey, err := bech32Decode(address)
	if publicKey != nil {
		publicKey = publicKey[:len(publicKey)-1]
		return publicKey, err
	} else {
		return nil, err
	}
}

//Apparel
func bech32Decode(encoded string) (string, []byte, error) {
	hrp, decoded, err := Decode(encoded)
	if err != nil {
		return hrp, decoded, err
	}

	data, err := ConvertBits(decoded, 5, 8, true)
	if err != nil {
		return hrp, data, err
	}

	return hrp, data, nil
}

func bech32Encode(hrp string, data []byte) (string, error) {
	conv, err := ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", err
	}

	encoded, err := Encode(hrp, conv)
	if err != nil {
		return "", err
	}

	return encoded, nil
}
