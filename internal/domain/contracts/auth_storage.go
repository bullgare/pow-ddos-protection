package contracts

import (
	"context"
)

// TODO ideally, some parameter (like bitLen in our case) could be dynamic, depending on traffic, for instance.
type AuthData struct {
	Seed   string
	UserID string
}

type AuthStorage interface {
	Store(ctx context.Context, data AuthData) error
	Delete(ctx context.Context, data AuthData) error
	CheckExists(ctx context.Context, data AuthData) (bool, error)
}
