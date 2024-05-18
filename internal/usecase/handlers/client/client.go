package client

import (
	"context"
	"fmt"
	"time"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

const numberOfRequests = 30

type Handler func(ctx context.Context)

func RunWordOfWisdom(
	authGenerator contracts.Authorizer,
	clientWOW contracts.ClientWordOfWisdom,
	onError func(error),
	shareInfo func(string),
) Handler {
	runOneCycle := func(ctx context.Context, iteration int) {
		// requesting auth params.
		authParams, err := clientWOW.GetAuthParams(ctx)
		if err != nil {
			onError(fmt.Errorf("%d: clientWOW.GetAuthParams: %w", iteration, err))
			return
		}

		// generating a token.
		start := time.Now()
		token, err := authGenerator.Generate(authParams.Seed)
		if err != nil {
			onError(fmt.Errorf("%d: generating token: %w", iteration, err))
			return
		}

		shareInfo(fmt.Sprintf("%d: using token %q (generated in %s)", iteration, token, time.Since(start).String()))

		reqData := contracts.DataRequest{
			OriginalSeed: authParams.Seed,
			Token:        token,
		}

		// requesting a genius quote.
		wowData, err := clientWOW.GetData(ctx, reqData)
		if err != nil {
			onError(fmt.Errorf("%d: clientWOW.GetData: %w", iteration, err))
			return
		}

		shareInfo(fmt.Sprintf("%d: success! got %q", iteration, wowData.MyPrecious))
	}

	return func(ctx context.Context) {
		for i := 0; i < numberOfRequests; i++ {
			select {
			case <-ctx.Done():
				onError(ctx.Err())
				return
			default:
				runOneCycle(ctx, i+1)
			}
		}
	}
}
