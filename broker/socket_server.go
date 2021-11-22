package broker

import (
	"context"
	"fmt"
	"net"
	"strings"

	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/storage"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vega/events"
	"golang.org/x/sync/errgroup"

	"github.com/golang/protobuf/proto"
	mangos "go.nanomsg.org/mangos/v3"
	mangosErr "go.nanomsg.org/mangos/v3/errors"
	"go.nanomsg.org/mangos/v3/protocol"
	"go.nanomsg.org/mangos/v3/protocol/pull"
	_ "go.nanomsg.org/mangos/v3/transport/inproc"
	_ "go.nanomsg.org/mangos/v3/transport/tcp"
)

const defaultEventChannelBufferSize = 256

// socketServer receives events from a remote broker.
// This is used by the data node to receive events from a non-validating core node.
type socketServer struct {
	log    *logging.Logger
	config *SocketConfig

	sock      protocol.Socket
	chainInfo storage.ChainInfoI
}

func pipeEventToString(pe mangos.PipeEvent) string {
	switch pe {
	case mangos.PipeEventAttached:
		return "Attached"
	case mangos.PipeEventDetached:
		return "Detached"
	default:
		return "Attaching"
	}
}

func newSocketServer(log *logging.Logger, config *SocketConfig, chainInfo storage.ChainInfoI) (*socketServer, error) {
	sock, err := pull.NewSocket()
	if err != nil {
		return nil, fmt.Errorf("failed to create new socket: %w", err)
	}

	return &socketServer{
		log:       log,
		config:    config,
		sock:      sock,
		chainInfo: chainInfo,
	}, nil
}

func (s socketServer) listen() error {
	addr := fmt.Sprintf(
		"%s://%s",
		strings.ToLower(s.config.TransportType),
		net.JoinHostPort(s.config.IP, fmt.Sprintf("%d", s.config.Port)),
	)

	if err := s.sock.Listen(addr); err != nil {
		return fmt.Errorf("failed to listen on %v: %w", addr, err)
	}

	s.log.Info("Starting broker socket server", logging.String("addr", s.config.IP), logging.Int("port", s.config.Port))

	s.sock.SetPipeEventHook(func(pe mangos.PipeEvent, p mangos.Pipe) {
		s.log.Info(
			"New broker connection event",
			logging.String("eventType", pipeEventToString(pe)),
			logging.Uint32("id", p.ID()),
			logging.String("address", p.Address()),
		)
	})

	return nil
}

func (s socketServer) receive(ctx context.Context) (<-chan events.Event, <-chan error) {
	streamStartReceived := false
	receiveCh := make(chan events.Event, defaultEventChannelBufferSize)
	var preStartEvents []events.Event
	errCh := make(chan error, 1)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		if err := s.close(); err != nil {
			return fmt.Errorf("failed to close socket: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		var recvTimeouts int

		for {
			var be eventspb.BusEvent

			msg, err := s.sock.Recv()
			if err != nil {
				switch err {
				case mangosErr.ErrRecvTimeout:
					s.log.Warn("Receive socket timeout", logging.Error(err))
					recvTimeouts++
					if recvTimeouts > s.config.MaxReceiveTimeouts {
						return fmt.Errorf("more then a 3 socket timeouts occurred: %w", err)
					}
				case mangosErr.ErrBadVersion:
					return fmt.Errorf("failed with bad protocol version: %w", err)
				case mangosErr.ErrClosed:
					return nil
				default:
					s.log.Error("Failed to receive message", logging.Error(err))
					continue
				}
			}

			if err := proto.Unmarshal(msg, &be); err != nil {
				s.log.Error("Failed to unmarshal received event", logging.Error(err))
				continue
			}

			if be.Version != eventspb.Version {
				return fmt.Errorf("mismatched BusEvent version received: %d, want %d", be.Version, eventspb.Version)
			}

			evt := toEvent(ctx, &be)
			if evt == nil {
				s.log.Error("Can not convert proto event to internal event", logging.String("event_type", be.GetType().String()))
				continue
			}

			streamStart, ok := evt.(*events.StreamStart)
			if ok {
				ourChainId, err := s.chainInfo.GetChainID()
				if err != nil {
					return fmt.Errorf("Unable to get expected chain ID %w", err)
				}

				if ourChainId == "" {
					// An empty chain ID indicates this is our first run
					ourChainId = streamStart.ChainId()
					s.chainInfo.SetChainID(ourChainId)
				}

				if streamStart.ChainId() != ourChainId {
					return fmt.Errorf("mismatched chain id received: %s, want %s", streamStart.ChainId(), ourChainId)
				}

				s.log.Infof("Stream starting with chain id: %s", streamStart.ChainId())
				// Send all the events we were queueing up before we got our StreamStart message
				for _, preStartEvent := range preStartEvents {
					receiveCh <- preStartEvent
				}
				preStartEvents = nil
				streamStartReceived = true

			}

			if streamStartReceived {
				receiveCh <- evt
			} else {
				preStartEvents = append(preStartEvents, evt)
			}

			recvTimeouts = 0
		}
	})

	go func() {
		defer func() {
			close(receiveCh)
			close(errCh)
		}()

		if err := eg.Wait(); err != nil {
			errCh <- err
		}
	}()

	return receiveCh, errCh
}

func (s socketServer) close() error {
	s.log.Info("Closing socket server")
	return s.sock.Close()
}
