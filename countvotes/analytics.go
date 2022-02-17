package countvotes

import (
	"encoding/json"
	"fmt"
	"results-analytics/client"

	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/crypto/nacl"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/vochain/scrutinizer/indexertypes"
	"go.vocdoni.io/proto/build/go/models"
)

func CountTargetVotes(client *client.Client, processID string, indexes []int, blankValue int) (int, int) {
	log.Infof("counting votes for process %s", processID)
	process, envelopes, err := getVotes(client, processID)
	if err != nil {
		log.Fatal(err)
	}
	var votes []*indexertypes.VotePackage
	// If process is encrypted, we need to decrypt each envelope
	if process.Envelope.EncryptedVotes {
		privKeys, err := client.GetProcessPrivKeys(process.ID)
		if err != nil {
			log.Fatal(err)
		}
		for _, envelope := range envelopes {
			vote, err := decryptVoteEnvelope(process, privKeys, envelope)
			if err != nil {
				log.Fatal(err)
			}
			votes = append(votes, vote)
		}
	} else {
		for _, envelope := range envelopes {
			vote := &indexertypes.VotePackage{}
			if err := json.Unmarshal(envelope.VotePackage, vote); err != nil {
				log.Fatalf("cannot unmarshal vote: %v", err)
			}
			votes = append(votes, vote)
		}
	}
	numBlankVotes := 0
voteLoop:
	for _, vote := range votes {
		for _, idx := range indexes {
			if len(vote.Votes) <= idx {
				log.Errorf("vote has only %d options, expected at least %d", len(vote.Votes), idx)
				continue voteLoop
			} else if vote.Votes[idx] != blankValue {
				// If any of the votes with the given indexes are not blank, go to the next vote loop
				continue voteLoop
			}
		}
		numBlankVotes++
	}
	return len(votes), numBlankVotes
}

func decryptVoteEnvelope(process *indexertypes.Process, keys []api.Key, envelope *indexertypes.EnvelopePackage) (*indexertypes.VotePackage, error) {
	// If package is encrypted
	if process.Status != int32(models.ProcessStatus_RESULTS.Number()) {
		return nil, fmt.Errorf("cannot decrypt envelope: process status is %d", process.Status)
	}
	if len(keys) < len(envelope.EncryptionKeyIndexes) {
		return nil, fmt.Errorf("cannot decrypt envelope: %d keys expected, %d provided",
			len(envelope.EncryptionKeyIndexes), len(keys))
	}
	var privKeys []string
	// Get the keys from each privKey with a matching index
	for _, idx := range envelope.EncryptionKeyIndexes {
		for _, key := range keys {
			if key.Idx == int(idx) {
				privKeys = append(privKeys, key.Key)
			}
		}
	}

	return unmarshalVote(envelope.VotePackage, privKeys)
}

func unmarshalVote(votePackage []byte, keys []string) (*indexertypes.VotePackage, error) {
	var vote indexertypes.VotePackage
	// if encryption keys, decrypt the vote
	if len(keys) > 0 {
		for i := len(keys) - 1; i >= 0; i-- {
			priv, err := nacl.DecodePrivate(keys[i])
			if err != nil {
				return nil, err
			}
			if votePackage, err = priv.Decrypt(votePackage); err != nil {
				return nil, fmt.Errorf("could not decrypt vote package: %v", err)
			}
		}
	}
	if err := json.Unmarshal(votePackage, &vote); err != nil {
		return nil, fmt.Errorf("cannot unmarshal vote: %w", err)
	}
	return &vote, nil
}
