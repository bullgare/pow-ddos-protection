package users

import (
	"time"
)

type User struct {
	RemoteAddress string
	RequestTime   time.Time
}
