package contract

import (
	"time"
)

type DataRequest struct {
	ClientRemoteAddress string
	RequestTime         time.Time
	Token               string
	OriginalSeed        string
}

type DataResponse struct {
	MyPrecious string
}
