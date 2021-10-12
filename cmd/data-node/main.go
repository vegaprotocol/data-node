package main

import (
	"context"
	"os"

	"code.vegaprotocol.io/data-node/cmd/data-node/cmds"
)

var (
	// VersionHash specifies the git commit used to build the application. See VERSION_HASH in Makefile for details.
	CLIVersionHash = ""

	// Version specifies the version used to build the application. See VERSION in Makefile for details.
	CLIVersion = ""
)

func main() {
	ctx := context.Background()
	if err := cmds.Main(ctx, CLIVersion, CLIVersionHash); err != nil {
		os.Exit(-1)
	}
}
