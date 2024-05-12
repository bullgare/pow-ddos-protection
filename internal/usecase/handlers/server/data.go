package server

import (
	"context"

	dcontracts "github.com/bullgare/pow-ddos-protection/internal/domain/contracts"
	ucontracts "github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

func Data(wowQuotes dcontracts.WOWQuotes) HandlerData {
	return func(ctx context.Context, req ucontracts.DataRequest) (ucontracts.DataResponse, error) {
		return ucontracts.DataResponse{
			MyPrecious: wowQuotes.GetRandomQuote(),
		}, nil
	}
}
