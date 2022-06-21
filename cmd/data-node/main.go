package main

import (
	"context"
	"fmt"
	"os"

	"code.vegaprotocol.io/data-node/config"
	"github.com/jessevdk/go-flags"
)

var (
	// VersionHash specifies the git commit used to build the application. See VERSION_HASH in Makefile for details.
	CLIVersionHash = ""

	// Version specifies the version used to build the application. See VERSION in Makefile for details.
	CLIVersion = ""
)

// Subcommand is the signature of a sub command that can be registered.
type Subcommand func(context.Context, *flags.Parser) error

// Register registers one or more subcommands.
func Register(ctx context.Context, parser *flags.Parser, cmds ...Subcommand) error {
	for _, fn := range cmds {
		if err := fn(ctx, parser); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	ctx := context.Background()
	if err := Main(ctx); err != nil {
		os.Exit(-1)
	}
}

func Main(ctx context.Context) error {
	parser := flags.NewParser(&config.Empty{}, flags.Default)

	if err := Register(ctx, parser,
		Init,
		Gateway,
		Node,
		Version,
		UnsafeResetAll,
	); err != nil {
		fmt.Printf("%+v\n", err)
		return err
	}

	if _, err := parser.Parse(); err != nil {
		switch t := err.(type) {
		case *flags.Error:
			if t.Type != flags.ErrHelp {
				parser.WriteHelp(os.Stdout)
			}
		}
		return err
	}
	return nil
}
