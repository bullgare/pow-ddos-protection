package server

import (
	"context"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

type HandlerAuth func(ctx context.Context, req contracts.AuthRequest) (contracts.AuthResponse, error)
type HandlerData func(ctx context.Context, req contracts.DataRequest) (contracts.DataResponse, error)
