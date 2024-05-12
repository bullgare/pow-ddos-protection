package connection

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"time"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/users"
)

// ProcessRequests handles all requests for one connection (meant to be used on a server side).
// TODO we might limit connection time to not exhaust the connections.
func (c *Connection) ProcessRequests(
	parentCtx context.Context,
	conn net.Conn,
	handler common.HandlerFunc,
) {
	defer func() { _ = conn.Close() }()

	if handler == nil {
		c.onError(errors.New("handler is required"))
		return
	}

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	for {
		select {
		case <-parentCtx.Done():
			return
		default:
			raw, err := r.ReadString(c.delimiter)
			if err == io.EOF {
				return
			}
			if err != nil {
				c.onError(fmt.Errorf("reading from conn: %w", err))
				continue
			}
			ctx := users.NewContext(parentCtx, users.User{
				RemoteAddress: conn.RemoteAddr().String(),
				RequestTime:   time.Now(),
			})

			msg, err := c.parseRawMessage(raw)
			if err != nil {
				c.onError(fmt.Errorf("parsing request: %w", err))
				c.sendResponse(w, common.Response{Type: common.MessageTypeError, Payload: []string{err.Error()}})
				continue
			}

			req := common.Request{
				Type:    msg.Type,
				Payload: slices.Clone(msg.Payload),
			}

			resp, err := handler(ctx, req)
			if err != nil {
				c.onError(err)
				c.sendResponse(w, common.Response{Type: common.MessageTypeError, Payload: []string{err.Error()}})
				continue
			}

			c.sendResponse(w, resp)
		}
	}
}

func (c *Connection) sendResponse(w *bufio.Writer, resp common.Response) {
	msg := common.Message{
		Version: common.MessageVersionV1,
		Type:    resp.Type,
		Payload: slices.Clone(resp.Payload),
	}

	err := c.sendMessage(w, msg)
	if err != nil {
		c.onError(err)
	}
}
