package connection

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
)

// SendRequest requests a server (meant to be used on a client side).
func (c *Connection) SendRequest(ctx context.Context, conn net.Conn, req common.Request) (common.Response, error) {
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	msg := common.Message{
		Version: common.MessageVersionV1,
		Type:    req.Type,
		Payload: slices.Clone(req.Payload),
	}

	err := c.sendMessage(w, msg)
	if err != nil {
		return common.Response{}, fmt.Errorf("sending request %v: %w", req, err)
	}

	select {
	case <-ctx.Done():
		return common.Response{}, ctx.Err()
	default:
		rawResp, err := r.ReadString(c.delimiter)
		if err != nil {
			return common.Response{}, fmt.Errorf("reading response: %w", err)
		}

		msg, err = c.parseRawMessage(rawResp)
		if err != nil {
			return common.Response{}, fmt.Errorf("parsing request: %w", err)
		}

		return common.Response{
			Type:    msg.Type,
			Payload: slices.Clone(msg.Payload),
		}, nil
	}
}
