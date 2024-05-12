package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"slices"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport"
)

func New(
	address string,
) (*Client, error) {
	if address == "" {
		return nil, errors.New("address is required")
	}

	return &Client{
		address: address,
	}, nil
}

type Client struct {
	address string
}

func (c *Client) SendRequest(ctx context.Context, req common.Request) (common.Response, error) {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return common.Response{}, fmt.Errorf("connecting to %q: %w", c.address, err)
	}
	defer func() { _ = conn.Close() }()

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	msg := common.Message{
		Version: common.MessageVersionV1,
		Type:    req.Type,
		Payload: slices.Clone(req.Payload),
	}

	err = transport.SendMessage(w, msg)
	if err != nil {
		return common.Response{}, fmt.Errorf("sending request %v: %w", req, err)
	}

	rawResp, err := r.ReadString(transport.MessageDelimiter)
	if err != nil {
		return common.Response{}, fmt.Errorf("reading response: %w", err)
	}

	msg, err = transport.ParseRawMessage(rawResp)
	if err != nil {
		return common.Response{}, fmt.Errorf("parsing request: %w", err)
	}

	return common.Response{
		Type:    msg.Type,
		Payload: slices.Clone(msg.Payload),
	}, nil
}
