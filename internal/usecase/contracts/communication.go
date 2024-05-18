package contracts

import (
	"context"
	"time"
)

type AuthRequest struct{}

type AuthResponse struct {
	Seed string
}

type DataRequest struct {
	ClientRemoteAddress string
	RequestTime         time.Time
	Token               string
	OriginalSeed        string
}

type DataResponse struct {
	MyPrecious string
}

type ClientWordOfWisdom interface {
	GetAuthParams(context.Context) (AuthResponse, error)
	GetData(context.Context, DataRequest) (DataResponse, error)
}
