//lint:file-ignore ST1003 Ignore underscores in names, this is straigh copied from the proto package to ease introducing the domain types

package types

import (
	commandspb "code.vegaprotocol.io/data-node/proto/vega/commands/v1"
)

type NodeSignature = commandspb.NodeSignature

type NodeSignatureKind = commandspb.NodeSignatureKind

const (
	// Represents an unspecified or missing value from the input
	NodeSignatureKind_NODE_SIGNATURE_KIND_UNSPECIFIED NodeSignatureKind = 0
	// Represents a signature for a new asset allow-listing
	NodeSignatureKind_NODE_SIGNATURE_KIND_ASSET_NEW NodeSignatureKind = 1
	// Represents a signature for an asset withdrawal
	NodeSignatureKind_NODE_SIGNATURE_KIND_ASSET_WITHDRAWAL NodeSignatureKind = 2
)
