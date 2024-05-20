package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/listener"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/server"
)

func New(
	lsn *listener.Listener,
	handlerAuth server.HandlerAuth,
	handlerData server.HandlerData,
	onError func(error),
) (*Server, error) {
	if lsn == nil {
		return nil, errors.New("listener is required")
	}
	if handlerAuth == nil {
		return nil, errors.New("usecase auth handler is required")
	}
	if handlerData == nil {
		return nil, errors.New("usecase data handler is required")
	}
	if onError == nil {
		return nil, errors.New("onError is required")
	}

	return &Server{
		lsn:         lsn,
		handlerAuth: handlerAuth,
		handlerData: handlerData,
		onError:     onError,
	}, nil
}

// Server knows how to handle incoming requests from transport/listener.
type Server struct {
	lsn         *listener.Listener
	handlerAuth server.HandlerAuth
	handlerData server.HandlerData
	onError     func(error)
}

func (s *Server) Start(ctx context.Context) error {
	return s.lsn.StartWithHandlerFunc(ctx, s.Handle)
}

func (s *Server) Stop() {
	s.lsn.Stop()
}

func (s *Server) Handle(ctx context.Context, req protocol.Request) (protocol.Response, error) {
	var (
		resp protocol.Response
		err  error
	)
	switch req.Type {
	case protocol.MessageTypeClientAuthReq:
		resp, err = s.Auth(ctx, req)
	case protocol.MessageTypeClientDataReq:
		resp, err = s.Data(ctx, req)
	default:
		resp, err = protocol.Response{},
			fmt.Errorf(
				"unexpected request to server: %q, only valid ones are: %s, %s",
				req.Type,
				protocol.MessageTypeClientAuthReq,
				protocol.MessageTypeClientDataReq,
			)
	}

	if err == nil {
		return resp, nil
	}

	s.onError(err)

	return protocol.Response{
		Type:    protocol.MessageTypeError,
		Payload: []string{err.Error()}, // TODO use status codes instead
	}, nil
}

func (s *Server) Auth(ctx context.Context, _ protocol.Request) (protocol.Response, error) {
	request := contracts.AuthRequest{}

	resp, err := s.handlerAuth(ctx, request)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("usecase auth handler for %v: %w", request, err)
	}

	return protocol.Response{
		Type:    protocol.MessageTypeSrvAuthResp,
		Payload: []string{resp.Seed},
	}, nil
}

func (s *Server) Data(ctx context.Context, req protocol.Request) (protocol.Response, error) {
	token, seed, err := protocol.MapPayloadToTokenAndSeed(req.Payload)
	if err != nil {
		return protocol.Response{}, err
	}

	request := contracts.DataRequest{
		Token:        token,
		OriginalSeed: seed,
	}

	resp, err := s.handlerData(ctx, request)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("usecase data handler for %v: %w", request, err)
	}

	return protocol.Response{
		Type:    protocol.MessageTypeSrvDataResp,
		Payload: []string{resp.Quote},
	}, nil
}
