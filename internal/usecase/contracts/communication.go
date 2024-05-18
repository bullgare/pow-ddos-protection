package contracts

import (
	"context"
)

type AuthRequest struct{}

type AuthResponse struct {
	Seed string
}

type DataRequest struct {
	Token        string
	OriginalSeed string
}

type DataResponse struct {
	MyPrecious string
}

type ClientWordOfWisdom interface {
	GetAuthParams(context.Context) (AuthResponse, error)
	GetData(context.Context, DataRequest) (DataResponse, error)
}
