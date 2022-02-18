package main

import (
	"fmt"
	"time"

	"results-analytics/client"
	"results-analytics/config"
	"results-analytics/countvotes"

	flag "github.com/spf13/pflag"
	"go.vocdoni.io/dvote/crypto/ethereum"
	"go.vocdoni.io/dvote/log"
)

func newConfig() (*config.Analytics, error) {
	cfg := config.Analytics{}
	// flags
	cfg.LogLevel = flag.String("logLevel", "info", "Log level (debug, info, warn, error, fatal)")
	cfg.SigningKey = flag.String("signingKey", "",
		"signing private Keys (if not specified, a new "+
			"one will be created), the first one is the oracle public key")
	cfg.GatewayUrl = flag.String("gatewayUrl",
		"https://gw1.vocdoni.net", "url to use as gateway api endpoint")
	cfg.ProcessID = flag.String("processID", "", "target process ID for analytics")
	cfg.QuestionIndexes = flag.IntSlice("questionIndexes", []int{0}, "array of indexes to check for target values")
	cfg.TargetValue = flag.Int("targetValue", 0, "value of a target vote")
	flag.CommandLine.SortFlags = false

	// parse flags
	flag.Parse()

	// Generate and save signing key if nos specified
	if len(*cfg.SigningKey) == 0 {
		fmt.Println("no signing keys, generating one...")
		signer := ethereum.NewSignKeys()
		signer.Generate()
		_, priv := signer.HexString()
		*cfg.SigningKey = priv
	}

	return &cfg, nil
}

func main() {
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Init(*cfg.LogLevel, "stdout")

	// Signer
	signer := ethereum.NewSignKeys()
	if err := signer.AddHexKey(*cfg.SigningKey); err != nil {
		log.Fatal(err)
	}
	pub, _ := signer.HexString()
	log.Infof("my public key: %s", pub)
	log.Infof("my address: %s", signer.AddressString())

	client, err := client.New(*cfg.GatewayUrl, signer)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("calculating number of votes for process %s with target value %d, indexes %v, gateway %s",
		*cfg.ProcessID, *cfg.TargetValue, *cfg.QuestionIndexes, *cfg.GatewayUrl)
	now := time.Now()
	totalVotes, targetVotes := countvotes.CountTargetVotes(client, *cfg.ProcessID, *cfg.QuestionIndexes, *cfg.TargetValue)
	log.Infof("of %d total votes for process %s, counted %d target votes\n TOOK %d seconds", totalVotes, *cfg.ProcessID, targetVotes, time.Now().Unix()-now.Unix())
}
