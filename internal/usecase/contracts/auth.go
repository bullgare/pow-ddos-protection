package contracts

import (
	"time"
)

// FIXME do we need it?
type AuthRequest struct {
	ClientRemoteAddress string
	RequestTime         time.Time
}

type AuthResponse struct {
	Seed string
}
