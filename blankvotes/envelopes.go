package blankvotes

import (
	"encoding/hex"
	"fmt"

	"results-analytics/client"

	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/vochain/scrutinizer/indexertypes"
	"go.vocdoni.io/proto/build/go/models"
)

func getVotes(client *client.Client, processID string) (*indexertypes.Process, []*indexertypes.EnvelopePackage, error) {
	pid, err := hex.DecodeString(processID)
	if err != nil {
		return nil, nil, err
	}
	process, err := client.GetProcess(pid)
	if err != nil {
		return nil, nil, err
	}

	if process.Envelope.EncryptedVotes && process.Status != int32(models.ProcessStatus_RESULTS) {
		return nil, nil, fmt.Errorf("process %x is encrypted and results are not published", pid)
	}

	// Fetch list of envelopes for process
	var envelopeList []*indexertypes.EnvelopeMetadata
	for {
		log.Infof("getting votes from %d", len(envelopeList))
		tempEnvelopeList, err := client.GetEnvelopeList(pid, len(envelopeList))
		if err != nil {
			return nil, nil, err
		}
		envelopeList = append(envelopeList, tempEnvelopeList...)
		if len(envelopeList) < 64 {
			break
		}
	}

	// Get the contents of each envelope
	var envelopes []*indexertypes.EnvelopePackage
	log.Infof("fetching individual vote envelopes")
	for _, envelope := range envelopeList {
		newEnvelope, err := client.GetEnvelope(envelope.Nullifier)
		if err != nil {
			return nil, nil, err
		}
		envelopes = append(envelopes, newEnvelope)
	}

	return process, envelopes, nil
}
