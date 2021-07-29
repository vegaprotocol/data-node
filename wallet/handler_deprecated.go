package wallet

import (
	"encoding/base64"

	"code.vegaprotocol.io/data-node/crypto"
	types "code.vegaprotocol.io/protos/vega"
	"github.com/golang/protobuf/proto"
)

func (h *Handler) SignTx(token, tx, pubKey string, blockHeight uint64) (SignedBundle, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// first the transaction would be in base64, let's decode
	rawTx, err := base64.StdEncoding.DecodeString(tx)
	if err != nil {
		return SignedBundle{}, err
	}

	kp, err := h.getKeyPair(token, pubKey)
	if err != nil {
		return SignedBundle{}, err
	}

	if kp.Tainted {
		return SignedBundle{}, ErrPubKeyIsTainted
	}

	txTy := &types.Transaction{
		InputData:   rawTx,
		Nonce:       crypto.NewNonce(),
		BlockHeight: blockHeight,
		From: &types.Transaction_PubKey{
			PubKey: kp.pubBytes,
		},
	}

	rawTxTy, err := proto.Marshal(txTy)
	if err != nil {
		return SignedBundle{}, err
	}

	// then lets sign the stuff and return it
	sig, err := kp.Algorithm.Sign(kp.privBytes, rawTxTy)
	if err != nil {
		return SignedBundle{}, err
	}

	return SignedBundle{
		Tx: rawTxTy,
		Sig: Signature{
			Sig:     sig,
			Algo:    kp.Algorithm.Name(),
			Version: kp.Algorithm.Version(),
		},
	}, nil
}
