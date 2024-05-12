package connection

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
)

func New(
	handler common.Handler,
	onError func(error),
) (*Connection, error) {
	if handler == nil {
		return nil, errors.New("conn is required")
	}
	if onError == nil {
		return nil, errors.New("onError is required")
	}

	return &Connection{
		handler: handler,
		onError: onError,
	}, nil
}

type Connection struct {
	handler common.Handler
	onError func(error)
}

// HandleBlocking handles all communications for one connection.
// TODO we might limit connection time to not exhaust the connections.
func (c *Connection) HandleBlocking(ctx context.Context, conn net.Conn) {
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	delim := []byte(common.Terminator)[0]

	for {
		select {
		case <-ctx.Done():
			return
		default:
			raw, err := r.ReadString(delim)
			if err == io.EOF {
				return
			}
			if err != nil {
				c.onError(fmt.Errorf("reading from conn: %w", err))
				continue
			}

			req, err := parseRequest(strings.TrimSpace(raw))
			if err != nil {
				c.onError(fmt.Errorf("parsing request: %w", err))
				c.sendResponse(w, common.Response{Type: common.MessageTypeError, Payload: []string{err.Error()}})
				continue
			}

			req.Meta.RemoteAddress = conn.RemoteAddr().String()

			resp, err := c.handler.Handle(ctx, req)
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
	stringResp := generateResponse(resp)

	_, err := w.WriteString(stringResp)
	if err != nil {
		c.onError(fmt.Errorf("sending response %q: %w", stringResp, err))
	}

	err = w.Flush()
	if err != nil {
		c.onError(fmt.Errorf("flushing response: %w", err))
	}
}

func parseRequest(raw string) (common.Request, error) {
	chunks := strings.Split(raw, common.Separator)
	if len(chunks) < 3 {
		return common.Request{}, fmt.Errorf("expected request to have at least 3 parts: version, message type and payload, got %d", len(chunks))
	}
	messageType := chunks[1]
	switch common.MessageType(messageType) {
	case common.MessageTypeError,
		common.MessageTypeClientAuthReq,
		common.MessageTypeSrvAuthResp,
		common.MessageTypeClientDataReq,
		common.MessageTypeSrvDataResp:
	default:
		return common.Request{}, fmt.Errorf("unexpected message type %q", messageType)
	}

	return common.Request{
		Type: common.MessageType(chunks[1]),
		Meta: common.RequestMeta{
			Time: time.Now(),
		},
		Payload: chunks[2:],
	}, nil
}

func generateResponse(resp common.Response) string {
	chunks := make([]string, 0, len(resp.Payload)+2)
	chunks = append(chunks,
		string(common.MessageVersionV1),
		string(resp.Type),
	)
	chunks = append(chunks, resp.Payload...)
	return strings.Join(chunks, common.Separator) + common.Terminator
}
