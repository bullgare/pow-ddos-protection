package listener

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/connection"
)

func New(
	address string,
	connHandler *connection.Connection,
	onError func(error),
) (*Listener, error) {
	if address == "" {
		return nil, errors.New("address is required")
	}
	if connHandler == nil {
		return nil, errors.New("connection handler is required")
	}
	if onError == nil {
		return nil, errors.New("onError is required")
	}

	chQuit := make(chan struct{})

	return &Listener{
		address:     address,
		connHandler: connHandler,
		onError:     onError,
		chQuit:      chQuit,
	}, nil
}

type Listener struct {
	address     string
	connHandler *connection.Connection
	onError     func(error)
	chQuit      chan struct{}
}

func (l *Listener) StartWithHandlerFunc(ctx context.Context, handler common.HandlerFunc) error {
	if handler == nil {
		return errors.New("handler is required")
	}

	lsn, err := net.Listen("tcp", l.address)
	if err != nil {
		return fmt.Errorf("listening %s: %w", l.address, err)
	}
	defer func() { _ = lsn.Close() }()

	l.handleConnections(ctx, lsn, handler)
	<-l.chQuit

	return nil
}

func (l *Listener) Stop() {
	close(l.chQuit)
}

func (l *Listener) handleConnections(ctx context.Context, lsn net.Listener, handler common.HandlerFunc) {
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				close(l.chQuit)
				return
			case <-l.chQuit:
				return
			default:
				conn, err := lsn.Accept()
				if err != nil {
					l.onError(fmt.Errorf("accepting tcp connection: %w", err))
					continue
				}

				go l.connHandler.ProcessRequests(ctx, conn, handler)
			}
		}
	}()
}
