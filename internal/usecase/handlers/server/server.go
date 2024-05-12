package server

import (
	"context"

	contract2 "github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/contract"
)

type HandlerAuth func(ctx context.Context, req contract2.AuthRequest) (contract2.AuthResponse, error)
type HandlerData func(ctx context.Context, req contract2.DataRequest) (contract2.DataResponse, error)
