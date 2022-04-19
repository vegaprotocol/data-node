package entities

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
)

type ERC20MultiSigSignerRemovedID struct{ ID }

func NewERC20MultiSigSignerRemovedID(id string) ERC20MultiSigSignerRemovedID {
	return ERC20MultiSigSignerRemovedID{ID: ID(id)}
}

type ERC20MultiSigSignerRemoved struct {
	ID          ERC20MultiSigSignerRemovedID
	ValidatorID NodeID
	OldSigner   EthereumAddress
	Submitter   EthereumAddress
	Nonce       string
	Timestamp   time.Time
	EpochID     int64
}

func ERC20MultiSigSignerRemovedFromProto(e *eventspb.ERC20MultiSigSignerRemoved) ([]*ERC20MultiSigSignerRemoved, error) {
	ents := []*ERC20MultiSigSignerRemoved{}

	epochID, err := strconv.ParseInt(e.EpochSeq, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing epoch '%v': %w", e.EpochSeq, err)
	}
	for _, s := range e.SignatureSubmitters {
		ents = append(ents, &ERC20MultiSigSignerRemoved{
			ID:          NewERC20MultiSigSignerRemovedID(s.SignatureId),
			Submitter:   NewEthereumAddress(strings.TrimPrefix(s.Submitter, "0x")),
			OldSigner:   NewEthereumAddress(strings.TrimPrefix(e.OldSigner, "0x")),
			ValidatorID: NewNodeID(e.ValidatorId),
			Nonce:       e.Nonce,
			Timestamp:   time.Unix(0, e.Timestamp),
			EpochID:     epochID,
		},
		)
	}

	return ents, nil
}
