package config

import (
	"log"
	"node/crypt"
	"os"

	"github.com/joho/godotenv"
)

var (
	// Node config
	NodeNdAddress    string
	NodeUwAddress    string
	NodeSecretKey    []byte
	ApiPort          string     // port for api server
	Ip               string     // node ip address
	BlockHeight      int64  = 1 // default block height
	GenesisAddress          = "uw194sxltwdyyyuznj8536h3c5hep8s4wqtdcqd9nm0w73ju4nq9nkq26fjyz"
	GenesisSecretKey        = []byte{21, 26, 239, 254, 223, 88, 200, 82, 118, 174, 63, 112, 154, 32, 83, 19, 203, 254, 207, 252, 206, 228, 78, 210, 146, 83, 50, 50, 139, 1, 50, 131, 45, 96, 111, 173, 205, 33, 9, 193, 78, 71, 164, 117, 120, 226, 151, 200, 79, 10, 184, 11, 110, 0, 210, 207, 111, 119, 163, 46,
		86, 96, 44, 236}
	BlockWriteTime                 = []int64{1, 11, 40, 51}
	Tax                    float64 = 0.001
	StorageMemoryLifeIter  int64   = 30       // blocks
	MaxStorageMemory       int64   = 50       // transactions
	RequestsMemoryLifeTime int64   = 60       // seconds
	Version                string  = "v1.4.4" // build version

	// Reward config
	RewardCoefficientStage1 float64 = 1.038107 / ((60 * 60 * 24 / 6) / float64(CalculateBlockWriteTime()))
	RewardCoefficientStage2 float64 = 15
	AnnualBlockHeight       int64   = 382763
	RewardTokenLabel        string  = "uwm"
	BaseToken               string  = "uwm"

	// Json Download for validators only
	JsonDownloadIp = ""

	// First peer data for not validators only
	FirstPeerIdx     int64  = 1
	FirstPeerIp      string = "185.151.245.69:65355"
	FirstPeerAddress string = "nd12l5h6aaza5mn39ayg79g29n75esp0wwjsuf2zg293gdpuc5rrmpq42mpa0"

	// API config
	ExplorerLimit            int64 = 30 // block limit for output to explorer
	BalanceTransactionsLimit int64 = 30 // transactions limit for output to wallet

	// Token config
	NewTokenCost1             float64 = 1
	NewTokenCost2             float64 = 5
	RenameTokenCost           float64 = 1
	MaxLabel                  int64   = 5
	MinLabel                  int64   = 3
	MaxName                   int64   = 80
	MinEmission               float64 = 10000000
	MaxEmission               float64 = 1000000000
	ChangeTokenStandardCost   float64 = 1
	FillTokenCardCost         float64 = 1
	FillTokenStandardCardCost float64 = 1

	TaxConversion float64 = 1

	// Smart-contracts config
	// Delegate sc config
	DelegateScAddress   string = "sc12l5h6aaza5mn39ayg79g29n75esp0wwjsuf2zg293gdpuc5rrmpq8k8vxh"
	DelegateBlockHeight int64  = 6
	DelegateToken       string = "uwm"
)

// function for calculating block write time
func CalculateBlockWriteTime() int64 {
	return (BlockWriteTime[3] - BlockWriteTime[2]) + (BlockWriteTime[2] - BlockWriteTime[1]) + (BlockWriteTime[1] - BlockWriteTime[0])
}

// function to load config data from config.env file
func Init() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Fatal("Error loading config.env file.")
	} else {
		if nodeMnemonic := os.Getenv("NODE_MNEMONIC"); nodeMnemonic == "" {
			log.Fatal("NODE_MNEMONIC is empty.")
		} else {
			NodeNdAddress = crypt.NodeAddressFromMnemonic(nodeMnemonic)
			NodeUwAddress = crypt.AddressFromMnemonic(nodeMnemonic)
			NodeSecretKey = crypt.SecretKeyFromSeed(crypt.SeedFromMnemonic(nodeMnemonic))
		}
		if apiPort := os.Getenv("API_PORT"); apiPort == "" {
			log.Fatal("API_PORT is empty.")
		} else {
			ApiPort = apiPort
		}
		if nodeIp := os.Getenv("NODE_IP"); nodeIp == "" {
			log.Fatal("NODE_IP is empty.")
		} else {
			Ip = nodeIp + ":65355"
		}
	}
}
