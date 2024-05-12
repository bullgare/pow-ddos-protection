package server

import (
	"context"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

func Data() HandlerData {
	return func(ctx context.Context, req contracts.DataRequest) (contracts.DataResponse, error) {
		return contracts.DataResponse{
			MyPrecious: "TODO", // TODO
		}, nil
	}
}
