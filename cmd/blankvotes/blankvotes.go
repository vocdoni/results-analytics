package main

import (
	"flag"
	"fmt"

	"github.com/vocdoni/results-analytics/config"
	"go.vocdoni.io/api/vocclient"
	"go.vocdoni.io/dvote/crypto/ethereum"
	"go.vocdoni.io/dvote/log"
)

func newConfig() (*config.Analytics, error) {
	var err error
	cfg := config.Analytics{}
	// flags
	cfg.LogLevel = *flag.String("logLevel", "info", "Log level (debug, info, warn, error, fatal)")
	cfg.SigningKey = *flag.String("signingKey", "",
		"signing private Keys (if not specified, a new "+
			"one will be created), the first one is the oracle public key")
	cfg.GatewayUrl = *flag.String("gatewayUrl",
		"https://api-dev.vocdoni.net", "url to use as gateway api endpoint")

	// parse flags
	flag.Parse()

	// Generate and save signing key if nos specified
	if len(cfg.SigningKey) == 0 {
		fmt.Println("no signing keys, generating one...")
		signer := ethereum.NewSignKeys()
		signer.Generate()
		if err != nil {
			return cfg, fmt.Errorf("cannot generate signing key: %v", err)
		}
		_, priv := signer.HexString()
		cfg.SigningKey = priv
	}

	return cfg, nil
}

func main() {
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}
	vocclient.New("", "")
}
