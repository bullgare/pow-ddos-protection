package client

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/connection"
)

func New(
	address string,
	connHandler *connection.Connection,
) (*Client, error) {
	if address == "" {
		return nil, errors.New("address is required")
	}
	if connHandler == nil {
		return nil, errors.New("connection handler is required")
	}

	return &Client{
		address:     address,
		connHandler: connHandler,
	}, nil
}

type Client struct {
	address     string
	connHandler *connection.Connection
}

func (c Client) SendRequest(ctx context.Context, req common.Request) (common.Response, error) {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return common.Response{}, fmt.Errorf("connecting to %q: %w", c.address, err)
	}
	defer func() { _ = conn.Close() }()

	return c.connHandler.SendRequest(ctx, conn, req)
}
