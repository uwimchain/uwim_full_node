package config

import (
	"log"
	"node/crypt"
	"os"

	"github.com/joho/godotenv"
)

var (
	NodeNdAddress    string                                                                   //
	NodeUwAddress    string                                                                   //
	NodeSecretKey    []byte                                                                   //
	ApiPort          string                                                                   // port for api server
	Ip               string                                                                   // node ip address
	BlockHeight      int64  = 1                                                               // default block height
	GenesisAddress          = "uw194sxltwdyyyuznj8536h3c5hep8s4wqtdcqd9nm0w73ju4nq9nkq26fjyz" //
	GenesisSecretKey        = []byte{21, 26, 239, 254, 223, 88, 200, 82, 118, 174, 63, 112, 154, 32, 83, 19, 203, 254,
		207, 252, 206, 228, 78, 210, 146, 83, 50, 50, 139, 1, 50, 131, 45, 96, 111, 173, 205, 33, 9, 193, 78, 71, 164,
		117, 120, 226, 151, 200, 79, 10, 184, 11, 110, 0, 210, 207, 111, 119, 163, 46, 86, 96, 44, 236} //
	BlockWriteTime                = []int64{1, 11, 40, 51} //
	Tax                   float64 = 0.001                  //
	StorageMemoryLifeIter int64   = 30                     // blocks
	MaxStorageMemory      int64   = 50                     // transactions
	Version               string  = "v1.6.1"               // build version
	MainNodeAddress       string  = "nd1kdpx2cyjq8mhk46asxz88jnqgnf4e0z03ywfcjdx49axqxpdpdvs38ngw9"

	RewardCoefficientStage1 float64 = 1.038107 / ((60 * 60 * 24 / 6) / float64(CalculateBlockWriteTime())) //
	EmitRate                        = []float64{0.350, 0.266, 0.207, 0.165, 0.136, 0.114, 0.098, 0.086, 0.077, 0.071,
		0.067, 0.064, 0.062, 0.059, 0.057, 0.054, 0.052, 0.050, 0.048, 0.046, 0.045, 0.043, 0.042, 0.041, 0.040}
	BlockAfterCharge = []int64{349000, 1349000, 2349000, 3349000, 4349000, 5349000, 6349000, 7349000, 8349000, 9349000,
		10349000, 11349000, 12349000, 13349000, 14349000, 15349000, 16349000, 17349000, 18349000, 19349000, 20349000,
		21349000, 22349000, 23349000, 24349000}
	AnnualBlockHeight int64  = 349000 //
	RewardTokenLabel  string = "uwm"  //
	BaseToken         string = "uwm"  //

	JsonDownloadIp string = ""

	FirstPeerIdx     int64  = 1
	FirstPeerIp      string = "194.88.107.114:65355"
	FirstPeerAddress string = "nd12l5h6aaza5mn39ayg79g29n75esp0wwjsuf2zg293gdpuc5rrmpq42mpa0"

	BalanceTransactionsLimit int   = 30 // transactions limit for output to wallet
	RequestsMemoryLifeTime   int64 = 1  // seconds
	RequestsMemoryCount      int64 = 5  // requests in memory count

	RenameTokenCost           float64 = 1                // "uwm"
	MaxLabel                  int64   = 5                //
	MinLabel                  int64   = 3                //
	MaxName                   int64   = 80               //
	MinEmission               float64 = 10000000         //
	MaxEmission               float64 = 1000000000       //
	FillTokenConfigCost       float64 = 1                // "uwm"
	TokenTypes                        = []int64{0, 1, 2} //
	NftTokenMaxCommission     float64 = 30               // percents
	NftTokenMinCommission     float64 = 0                // percents

	TaxConversion float64 = 1

	DelegateScAddress string  = "sc12l5h6aaza5mn39ayg79g29n75esp0wwjsuf2zg293gdpuc5rrmpq8k8vxh" //
	DelegateToken     string  = "uwm"                                                           //
	DelegateEmitRate  float64 = 0.6

	HolderScAddress string = "sc1cq56xp5hfks3slpcneaty2r2mnjk28q6ea7e7dw6s7zcl96j2scq4vvfkf" //

	VoteScAddress               string = "sc1ksu3zurrs7mgl844wx6unchzwpuhqlhvn9hh2zcp2xvs46g0sx6qh2pq7r" //
	VoteSuperAddress            string = "uw128za3mp36cdzwp5ttr268zx5tmvrs40h639qljystxhc0nsy9zkqxyclqy" //
	VoteAnswerOptionDefaultCost        = 0.5                                                             // "uwm"
	MaxVoteMemory               int    = 10                                                              //
	MaxVoteAnswerOptions        int    = 10                                                              //

	NftCreateCost             float64 = 0.1   // "uwm"
	NftTokenElsCountMax       int     = 10000 // token elements
	NftTokenElMaxDataFieldLen int     = 8000  // symbols
	NftTokenElCreateLimit     int     = 20    // token elements
)

func CalculateBlockWriteTime() int64 {
	return (BlockWriteTime[3] - BlockWriteTime[2]) + (BlockWriteTime[2] - BlockWriteTime[1]) + (BlockWriteTime[1] - BlockWriteTime[0])
}

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

func GetEmitRateIdx() int {
	for i := range BlockAfterCharge {
		if i+1 == len(BlockAfterCharge) {
			return i
		}

		if BlockHeight >= BlockAfterCharge[i] && BlockHeight <= BlockAfterCharge[i+1] {
			return i
		}
	}

	return -1
}