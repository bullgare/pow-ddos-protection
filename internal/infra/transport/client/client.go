package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport"
)

const (
	dialTimeout       = 100 * time.Millisecond
	connectionTimeout = 500 * time.Millisecond
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

func (c *Client) SendRequest(ctx context.Context, req protocol.Request) (protocol.Response, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := (&net.Dialer{Timeout: dialTimeout}).DialContext(ctx, "tcp", c.address)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("connecting to %q: %w", c.address, err)
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(connectionTimeout))
	go func() {
		<-ctx.Done()
		_ = conn.SetDeadline(time.Now())
	}()

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	msg := protocol.Message{
		Version: protocol.MessageVersionV1,
		Type:    req.Type,
		Payload: slices.Clone(req.Payload),
	}

	err = transport.SendMessage(w, msg)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("sending request %v: %w", req, err)
	}

	rawResp, err := r.ReadString(transport.MessageDelimiter)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("reading response: %w", err)
	}

	msg, err = transport.ParseRawMessage(rawResp)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("parsing response: %w", err)
	}

	return protocol.Response{
		Type:    msg.Type,
		Payload: slices.Clone(msg.Payload),
	}, nil
}
