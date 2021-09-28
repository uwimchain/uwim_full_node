package crypt

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"golang.org/x/crypto/pbkdf2"
	"math"
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

func SignMessageWithSecretKey(secretKey []byte, message []byte) []byte {
	return ed25519.Sign(secretKey, message)
}

func VerifySign(publicKey []byte, data []byte, signature []byte) bool {
	return ed25519.Verify(publicKey, data, signature)
}

func GetHash(jsonString []byte) string {
	alg := sha256.New()
	alg.Write(jsonString)
	return hex.EncodeToString(alg.Sum(nil))
}

func SeedFromMnemonic(mnemonic string) []byte {
	return pbkdf2.Key([]byte(mnemonic), nil, 2048, 32, sha512.New)
}

func SecretKeyFromSeed(seed []byte) []byte {
	return ed25519.NewKeyFromSeed(seed)
}

func PublicKeyFromSecretKey(secretKey []byte) []byte {
	publicKey := make([]byte, ed25519.PublicKeySize)
	copy(publicKey, secretKey[32:])

	return publicKey
}

func DecodeTransactionRaw(transactionRaw string) (string, string, string, float64, string, []byte) {
	tx, _ := base64.StdEncoding.DecodeString(transactionRaw)

	commentTitleBytes := tx[:1]
	commentTitle := ""
	switch commentTitleBytes[0] {
	case 1:
		commentTitle = "default_transaction"
		break
	}

	senderBytes := tx[1:35]

	senderPrefixBytes := senderBytes[:2]
	senderPrefix := getPrefixForBytes(senderPrefixBytes)
	senderAddressBytes := senderBytes[2:]

	senderAddress, _ := bech32Encode(senderPrefix, senderAddressBytes)

	recipientBytes := tx[35:69]


	recipientPrefixBytes := recipientBytes[:2]
	recipientPrefix := getPrefixForBytes(recipientPrefixBytes)
	recipientAddressBytes := recipientBytes[2:]

	recipientAddress, _ := bech32Encode(recipientPrefix, recipientAddressBytes)

	amountBytes := tx[69:86]
	amountFirstBytes := amountBytes[:8]
	amountSecondBytes := amountBytes[8:16]
	countZerosAfterDotBytes := amountBytes[16:]

	amountFirst := float64(binary.BigEndian.Uint64(amountFirstBytes))
	amountSecond := float64(binary.BigEndian.Uint64(amountSecondBytes))
	countZerosAfterDot := float64(countZerosAfterDotBytes[0]) + 1

	amountSecond /= math.Pow(10, countZerosAfterDot)
	amount := amountFirst + amountSecond

	tokenBytes := tx[86:120]
	tokenScAddressBytes := tokenBytes[2:]

	tokenScAddress, _ := bech32Encode(metrics.SmartContractPrefix, tokenScAddressBytes)

	signatureBytes := tx[120:]

	return commentTitle, senderAddress, recipientAddress, amount, tokenScAddress, signatureBytes
}

func getPrefixForBytes(prefixBytes []byte) string {
	switch string(prefixBytes) {
	case string([]byte{86, 224}):
		return "uw"
	case string([]byte{76, 96}):
		return "sc"
	case string([]byte{56, 128}):
		return "nd"
	default:
		return ""
	}
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
