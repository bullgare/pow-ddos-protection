package contract

import (
	"time"
)

type AuthRequest struct {
	ClientRemoteAddress string
	RequestTime         time.Time
}

type AuthResponse struct {
	Seed string
}
