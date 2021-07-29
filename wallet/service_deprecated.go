package wallet

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"

	api "code.vegaprotocol.io/protos/data-node/api"
	typespb "code.vegaprotocol.io/protos/vega"
	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/grpc/status"
)

func (s *Service) SignTxSync(t string, w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	s.signTx(t, w, r, p, api.SubmitTransactionRequest_TYPE_SYNC)
}

func (s *Service) SignTxCommit(t string, w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	s.signTx(t, w, r, p, api.SubmitTransactionRequest_TYPE_COMMIT)
}

func (s *Service) SignTx(t string, w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	s.signTx(t, w, r, p, api.SubmitTransactionRequest_TYPE_ASYNC)
}

func (s *Service) signTx(t string, w http.ResponseWriter, r *http.Request, _ httprouter.Params, ty api.SubmitTransactionRequest_Type) {
	req := SignTxRequest{}
	if err := unmarshalBody(r, &req); err != nil {
		writeError(w, newError(err.Error()), http.StatusBadRequest)
		return
	}
	if len(req.Tx) <= 0 {
		writeError(w, newError("missing tx field"), http.StatusBadRequest)
		return
	}
	if len(req.PubKey) <= 0 {
		writeError(w, newError("missing pubKey field"), http.StatusBadRequest)
		return
	}

	height, err := s.nodeClient.LastBlockHeight(r.Context())
	if err != nil {
		writeError(w, newError("could not get last block height"), http.StatusInternalServerError)
		return
	}

	sb, err := s.handler.SignTx(t, req.Tx, req.PubKey, height)
	if err != nil {
		writeError(w, newError(err.Error()), http.StatusForbidden)
		return
	}

	if req.Propagate {
		if err := s.nodeForward.Send(r.Context(), &sb, ty); err != nil {
			if s, ok := status.FromError(err); ok {
				details := []string{}
				for _, v := range s.Details() {
					v := v.(*typespb.ErrorDetail)
					details = append(details, v.Message)
				}
				writeError(w, newErrorWithDetails(err.Error(), details), http.StatusInternalServerError)
			} else {
				writeError(w, newError(err.Error()), http.StatusInternalServerError)
			}
			return
		}
	}

	rawBundle, err := proto.Marshal(sb.IntoProto())
	if err != nil {
		writeError(w, newError(err.Error()), http.StatusInternalServerError)
		return
	}

	hexBundle := hex.EncodeToString(rawBundle)
	base64Bundle := base64.StdEncoding.EncodeToString(rawBundle)

	res := SignTxResponse{
		SignedTx:     sb,
		HexBundle:    hexBundle,
		Base64Bundle: base64Bundle,
	}

	writeSuccess(w, res, http.StatusOK)
}
