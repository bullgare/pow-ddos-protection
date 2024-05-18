package contracts

import (
	"context"
)

type AuthData struct {
	Seed   string
	UserID string
}

type AuthStorage interface {
	Store(ctx context.Context, data AuthData) error
	Delete(ctx context.Context, data AuthData) error
	CheckExists(ctx context.Context, data AuthData) (bool, error)
}
