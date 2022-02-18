package countvotes

import (
	"encoding/hex"
	"fmt"
	"sync"

	"results-analytics/client"

	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/vochain/scrutinizer/indexertypes"
	"go.vocdoni.io/proto/build/go/models"
)

const maxConcurrentRequests = 128

func getVotes(client *client.Client, processID string) (*indexertypes.Process, []*indexertypes.EnvelopePackage, error) {
	pid, err := hex.DecodeString(processID)
	if err != nil {
		return nil, nil, err
	}
	process, err := client.GetProcess(pid)
	if err != nil {
		return nil, nil, err
	}

	if process.Envelope.EncryptedVotes && process.Status != int32(models.ProcessStatus_RESULTS) && process.Status != int32(models.ProcessStatus_ENDED) {
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
		if len(tempEnvelopeList) < 64 {
			break
		}
	}

	// Get the contents of each envelope
	var envelopes []*indexertypes.EnvelopePackage
	envelopeCh := make(chan *indexertypes.EnvelopePackage, 100000)
	wg := new(sync.WaitGroup)
	getEnvelope := func(nullifier types.HexBytes) {
		newEnvelope, err := client.GetEnvelope(nullifier)
		if err != nil {
			log.Error(err)
		}
		envelopeCh <- newEnvelope
		wg.Done()
	}

	envelopeIndex := 0
	for envelopeIndex < len(envelopeList) {
		// Only make maxConcurrentRequests requests, wait for them to finish before starting next batch
		log.Infof("fetching individual vote envelopes from %d", envelopeIndex)
		for i := 0; i < maxConcurrentRequests && envelopeIndex+i < len(envelopeList); i++ {
			wg.Add(1)
			go getEnvelope(envelopeList[envelopeIndex+i].Nullifier)
			envelopeIndex++
		}

		go func() {
			for {
				newEnvelope := <-envelopeCh
				envelopes = append(envelopes, newEnvelope)
			}
		}()

		wg.Wait()
	}

	return process, envelopes, nil
}
