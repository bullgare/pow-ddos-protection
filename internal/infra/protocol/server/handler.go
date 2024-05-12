package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
	contract2 "github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/contract"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/server"
)

func New(handlerAuth server.HandlerAuth, handlerData server.HandlerData, onError func(error)) (*Handler, error) {
	if handlerAuth == nil {
		return nil, errors.New("usecase auth handler is required")
	}
	if handlerData == nil {
		return nil, errors.New("usecase data handler is required")
	}
	if onError == nil {
		return nil, errors.New("onError is required")
	}

	return &Handler{
		handlerAuth: handlerAuth,
		handlerData: handlerData,
		onError:     onError,
	}, nil
}

var _ common.Handler = &Handler{}

type Handler struct {
	handlerAuth server.HandlerAuth
	handlerData server.HandlerData
	onError     func(error)
}

func (h *Handler) Handle(ctx context.Context, req common.Request) (common.Response, error) {
	var (
		resp common.Response
		err  error
	)
	switch req.Type {
	case common.MessageTypeClientAuthReq:
		resp, err = h.Auth(ctx, req)
	case common.MessageTypeClientDataReq:
		resp, err = h.Data(ctx, req)
	default:
		resp, err = common.Response{},
			fmt.Errorf(
				"unexpected request to server: %q, only valid ones are: %s, %s",
				req.Type,
				common.MessageTypeClientAuthReq,
				common.MessageTypeClientDataReq,
			)
	}

	if err == nil {
		return resp, nil
	}

	h.onError(err)

	return common.Response{
		Type:    common.MessageTypeError,
		Payload: []string{err.Error()}, // TODO use status codes instead
	}, nil
}

func (h *Handler) Auth(ctx context.Context, req common.Request) (common.Response, error) {
	request := contract2.AuthRequest{
		ClientRemoteAddress: req.Meta.RemoteAddress,
		RequestTime:         req.Meta.Time,
	}

	resp, err := h.handlerAuth(ctx, request)
	if err != nil {
		return common.Response{}, fmt.Errorf("usecase auth handler for %v: %w", request, err)
	}

	return common.Response{
		Type:    common.MessageTypeSrvAuthResp,
		Payload: []string{resp.Seed},
	}, nil
}

func (h *Handler) Data(ctx context.Context, req common.Request) (common.Response, error) {
	token, seed, err := common.MapPayloadToTokenAndSeed(req.Payload)
	if err != nil {
		return common.Response{}, err
	}

	request := contract2.DataRequest{
		ClientRemoteAddress: req.Meta.RemoteAddress,
		RequestTime:         req.Meta.Time,
		Token:               token,
		OriginalSeed:        seed,
	}

	resp, err := h.handlerData(ctx, request)
	if err != nil {
		return common.Response{}, fmt.Errorf("usecase data handler for %v: %w", request, err)
	}

	return common.Response{
		Type:    common.MessageTypeSrvDataResp,
		Payload: []string{resp.MyPrecious},
	}, nil
}
