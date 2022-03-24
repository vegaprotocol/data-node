package entities

import (
	"encoding/hex"
	"fmt"
	"time"

	oraclespb "code.vegaprotocol.io/protos/vega/oracles/v1"
)

type SpecID struct {
	ID
}

func NewSpecID(id string) SpecID {
	return SpecID{
		ID: ID(id),
	}
}

type PublicKey = []byte
type PublicKeyList = []PublicKey

type OracleSpec struct {
	ID         SpecID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	PublicKeys PublicKeyList
	Filters    []Filter
	Status     OracleSpecStatus
	VegaTime   time.Time
}

func OracleSpecFromProto(spec *oraclespb.OracleSpec, vegaTime time.Time) (*OracleSpec, error) {
	id := NewSpecID(spec.Id)
	pubKeys, err := decodePublicKeys(spec.PubKeys)
	if err != nil {
		return nil, err
	}

	filters := filtersFromProto(spec.Filters)

	return &OracleSpec{
		ID:         id,
		CreatedAt:  time.Unix(0, spec.CreatedAt),
		UpdatedAt:  time.Unix(0, spec.UpdatedAt),
		PublicKeys: pubKeys,
		Filters:    filters,
		Status:     OracleSpecStatus(spec.Status),
		VegaTime:   vegaTime,
	}, nil
}

func (os *OracleSpec) ToProto() *oraclespb.OracleSpec {
	pubKeys := make([]string, 0, len(os.PublicKeys))

	for _, pk := range os.PublicKeys {
		pubKey := hex.EncodeToString(pk)
		pubKeys = append(pubKeys, pubKey)
	}

	filters := filtersToProto(os.Filters)

	return &oraclespb.OracleSpec{
		Id:        os.ID.String(),
		CreatedAt: os.CreatedAt.UnixNano(),
		UpdatedAt: os.UpdatedAt.UnixNano(),
		PubKeys:   pubKeys,
		Filters:   filters,
		Status:    oraclespb.OracleSpec_Status(os.Status),
	}
}

func decodePublicKeys(publicKeys []string) (PublicKeyList, error) {
	pkList := make(PublicKeyList, 0, len(publicKeys))

	for _, publicKey := range publicKeys {
		pk, err := hex.DecodeString(publicKey)
		if err != nil {
			return nil, fmt.Errorf("cannot decode public key: %s", publicKey)
		}

		pkList = append(pkList, pk)
	}

	return pkList, nil
}
