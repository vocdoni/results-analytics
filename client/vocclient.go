package client

import (
	"fmt"

	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/client"
	"go.vocdoni.io/dvote/crypto/ethereum"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/vochain/scrutinizer/indexertypes"
)

type Client struct {
	gw         *client.Client
	signingKey *ethereum.SignKeys
}

// New initializes a new gatewayPool with the gatewayUrls, in order of health
// returns the new Client
func New(gatewayUrl string, signingKey *ethereum.SignKeys) (*Client, error) {
	gw, err := DiscoverGateway(gatewayUrl)
	if err != nil {
		return nil, err
	}

	return &Client{
		gw:         gw,
		signingKey: signingKey,
	}, nil
}

// ActiveEndpoint returns the address of the current active endpoint, if one exists
func (c *Client) ActiveEndpoint() string {
	if c.gw == nil {
		return ""
	}
	return c.gw.Addr
}

func (c *Client) request(req api.APIrequest,
	signer *ethereum.SignKeys) (*api.APIresponse, error) {
	resp, err := c.gw.Request(req, signer)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf(resp.Message)
	}
	return resp, nil
}

// GetProcess returns the process parameters for the given process id
func (c *Client) GetProcess(pid []byte) (*indexertypes.Process, error) {
	req := api.APIrequest{Method: "getProcessInfo", ProcessID: pid}
	resp, err := c.request(req, c.signingKey)
	if err != nil {
		return nil, err
	}
	if !resp.Ok || resp.Process == nil {
		return nil, fmt.Errorf("cannot getProcessInfo: %v", resp.Message)
	}
	if resp.Process.Metadata == "" {
		return nil, fmt.Errorf("election metadata not yet set")
	}
	return resp.Process, nil
}

// GetProcessKeys returns the encryption privKeys for a process
func (c *Client) GetProcessPrivKeys(pid []byte) ([]api.Key, error) {
	req := api.APIrequest{Method: "getProcessKeys", ProcessID: pid}
	resp, err := c.request(req, c.signingKey)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("could not get process keys: %s", resp.Message)
	}
	return resp.EncryptionPrivKeys, nil
}

// GetEnvelopeList returns a list of envelope metadata for the given process
func (c *Client) GetEnvelopeList(pid []byte, from int) ([]*indexertypes.EnvelopeMetadata, error) {
	req := api.APIrequest{
		Method:    "getEnvelopeList",
		From:      from,
		ListSize:  64,
		ProcessID: pid,
	}
	resp, err := c.request(req, c.signingKey)
	if err != nil {
		return nil, err
	}
	if !resp.Ok || resp.Envelopes == nil {
		return nil, fmt.Errorf("cannot getEnvelopeList: %v", resp.Message)
	}
	return resp.Envelopes, nil
}

// GetEnvelope returns the contents of a single vote envelope
func (c *Client) GetEnvelope(nullifier types.HexBytes) (*indexertypes.EnvelopePackage, error) {
	req := api.APIrequest{
		Method:    "getEnvelope",
		Nullifier: nullifier,
	}
	resp, err := c.request(req, c.signingKey)
	if err != nil {
		return nil, err
	}
	if !resp.Ok || resp.Envelope == nil {
		return nil, fmt.Errorf("cannot getEnvelope: %v", resp.Message)
	}
	return resp.Envelope, nil
}
