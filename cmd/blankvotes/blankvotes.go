package main

import (
	"fmt"

	"results-analytics/blankvotes"
	"results-analytics/client"
	"results-analytics/config"

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
	cfg.VoteIndexes = flag.IntSlice("voteIndexes", []int{0}, "array of indexes to check for blank values")
	cfg.BlankValue = flag.Int("blankValue", 0, "value of a 'blank' vote")
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
	log.Infof("calculating number of blank votes for process %s with blank value %d, indexes %v, gateway %s",
		*cfg.ProcessID, *cfg.BlankValue, *cfg.VoteIndexes, *cfg.GatewayUrl)
	totalVotes, blankVotes := blankvotes.CountBlankVotes(client, *cfg.ProcessID, *cfg.VoteIndexes, *cfg.BlankValue)
	log.Infof("of %d total votes for process %s, counted %d blank votes", totalVotes, *cfg.ProcessID, blankVotes)
}
