package client

import (
	"context"
	"fmt"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

const numberOfRequests = 3

type Handler func(ctx context.Context)

func RunWordOfWisdom(
	clientWOW contracts.ClientWordOfWisdom,
	onError func(error),
	shareInfo func(string),
) Handler {
	runOneCycle := func(ctx context.Context) {
		authParams, err := clientWOW.GetAuthParams(ctx)
		if err != nil {
			onError(fmt.Errorf("clientWOW.GetAuthParams: %w", err))
			return
		}

		// FIXME add pow auth logic
		token := ""
		shareInfo(fmt.Sprintf("using token %q", token))

		reqData := contracts.DataRequest{
			OriginalSeed: authParams.Seed,
			Token:        token,
		}

		wowData, err := clientWOW.GetData(ctx, reqData)
		if err != nil {
			onError(fmt.Errorf("clientWOW.GetData: %w", err))
			return
		}

		shareInfo(fmt.Sprintf("success! got %q", wowData.MyPrecious))
	}

	return func(ctx context.Context) {
		for i := 0; i < numberOfRequests; i++ {
			select {
			case <-ctx.Done():
				onError(ctx.Err())
				return
			default:
				runOneCycle(ctx)
			}
		}
	}
}
