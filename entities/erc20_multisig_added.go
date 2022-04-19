package entities

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
)

type ERC20MultiSigSignerAddedID struct{ ID }

func NewERC20MultiSigSignerAddedID(id string) ERC20MultiSigSignerAddedID {
	return ERC20MultiSigSignerAddedID{ID: ID(id)}
}

type ERC20MultiSigSignerAdded struct {
	ID          ERC20MultiSigSignerAddedID
	ValidatorID NodeID
	NewSigner   EthereumAddress
	Submitter   EthereumAddress
	Nonce       string
	Timestamp   time.Time
	EpochID     int64
}

func ERC20MultiSigSignerAddedFromProto(e *eventspb.ERC20MultiSigSignerAdded) (*ERC20MultiSigSignerAdded, error) {
	epochID, err := strconv.ParseInt(e.EpochSeq, 10, 64)
	if err != nil {
		return &ERC20MultiSigSignerAdded{}, fmt.Errorf("parsing epoch '%v': %w", e.EpochSeq, err)
	}
	return &ERC20MultiSigSignerAdded{
		ID:          NewERC20MultiSigSignerAddedID(e.SignatureId),
		ValidatorID: NewNodeID(e.ValidatorId),
		NewSigner:   NewEthereumAddress(strings.TrimPrefix(e.NewSigner, "0x")),
		Submitter:   NewEthereumAddress(strings.TrimPrefix(e.Submitter, "0x")),
		Nonce:       e.Nonce,
		Timestamp:   time.Unix(0, e.Timestamp),
		EpochID:     epochID,
	}, nil
}
