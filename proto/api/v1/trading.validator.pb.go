// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: api/v1/trading.proto

package v1

import (
	fmt "fmt"
	math "math"

	_ "code.vegaprotocol.io/data-node/proto/commands/v1"
	proto "github.com/golang/protobuf/proto"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *SubmitTransactionRequest) Validate() error {
	if this.Tx != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Tx); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Tx", err)
		}
	}
	return nil
}
func (this *SubmitTransactionResponse) Validate() error {
	return nil
}
func (this *PrepareWithdrawRequest) Validate() error {
	if this.Withdraw != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Withdraw); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Withdraw", err)
		}
	}
	return nil
}
