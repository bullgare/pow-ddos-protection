package server

import (
	"context"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/handlers/contract"
)

func Data() HandlerData {
	return func(ctx context.Context, req contract.DataRequest) (contract.DataResponse, error) {
		return contract.DataResponse{
			MyPrecious: "TODO", // TODO
		}, nil
	}
}
