package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/connection"
)

func New(
	address string,
	connHandler *connection.Connection,
	onError func(error),
) (*Server, error) {
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

	return &Server{
		address:     address,
		connHandler: connHandler,
		onError:     onError,
		chQuit:      chQuit,
	}, nil
}

type Server struct {
	address     string
	connHandler *connection.Connection
	onError     func(error)
	chQuit      chan struct{}
}

func (s *Server) Start(ctx context.Context) error {
	lsn, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("listening %s: %w", s.address, err)
	}
	defer lsn.Close()

	s.handleConnections(ctx, lsn)
	<-s.chQuit

	return nil
}

func (s *Server) Stop() {
	close(s.chQuit)
}

func (s *Server) handleConnections(ctx context.Context, lsn net.Listener) {
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				close(s.chQuit)
				return
			case <-s.chQuit:
				return
			default:
				conn, err := lsn.Accept() // FIXME check errors
				if err != nil {
					s.onError(fmt.Errorf("accepting tcp connection: %w", err))
					continue
				}

				go s.connHandler.HandleBlocking(ctx, conn)
			}
		}
	}()
}
