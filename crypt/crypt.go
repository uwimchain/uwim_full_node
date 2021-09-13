package crypt

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
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
