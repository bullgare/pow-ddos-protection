package contracts

import (
	"context"
)

type ClientWordOfWisdom interface {
	GetAuthParams(context.Context) (AuthResponse, error)
	GetData(context.Context, DataRequest) (DataResponse, error)
}
