package connection

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
)

// FIXME split it into 3 files.
func New(
	onError func(error),
) (*Connection, error) {
	if onError == nil {
		return nil, errors.New("onError is required")
	}

	return &Connection{
		onError:   onError,
		delimiter: []byte(common.Terminator)[0],
	}, nil
}

type Connection struct {
	onError   func(error)
	delimiter byte
}

// HandleBlockingWithHandlerFunc handles all communications for one connection (meant to be used on a server side).
// TODO we might limit connection time to not exhaust the connections.
func (c *Connection) HandleBlockingWithHandlerFunc(
	ctx context.Context,
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
		case <-ctx.Done():
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

			msg, err := parseRawMessage(raw)
			if err != nil {
				c.onError(fmt.Errorf("parsing request: %w", err))
				c.sendResponse(w, common.Response{Type: common.MessageTypeError, Payload: []string{err.Error()}})
				continue
			}

			req := common.Request{
				Type: msg.Type,
				Meta: common.RequestMeta{
					RemoteAddress: conn.RemoteAddr().String(),
					Time:          time.Now(),
				},
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

		msg, err = parseRawMessage(rawResp)
		if err != nil {
			return common.Response{}, fmt.Errorf("parsing request: %w", err)
		}

		return common.Response{
			Type:    msg.Type,
			Payload: slices.Clone(msg.Payload),
		}, nil
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

func (c *Connection) sendMessage(w *bufio.Writer, msg common.Message) error {
	raw := generateRawMessage(msg)

	_, err := w.WriteString(raw)
	if err != nil {
		return fmt.Errorf("sending response %q: %w", raw, err)
	}

	err = w.Flush()
	if err != nil {
		return fmt.Errorf("flushing response: %w", err)
	}

	return nil
}

func parseRawMessage(raw string) (common.Message, error) {
	raw = strings.TrimSpace(raw)
	chunks := strings.Split(raw, common.Separator)
	if len(chunks) < 3 {
		return common.Message{}, fmt.Errorf("expected message to have at least 3 parts: version, message type and payload, got %d", len(chunks))
	}

	version := common.MessageVersion(chunks[0])
	if version != common.MessageVersionV1 {
		return common.Message{}, fmt.Errorf("unexpected message version %s, should be %s", version, common.MessageVersionV1)
	}

	messageType := chunks[1]
	switch common.MessageType(messageType) {
	case common.MessageTypeError,
		common.MessageTypeClientAuthReq,
		common.MessageTypeSrvAuthResp,
		common.MessageTypeClientDataReq,
		common.MessageTypeSrvDataResp:
	default:
		return common.Message{}, fmt.Errorf("unexpected message type %q", messageType)
	}

	return common.Message{
		Version: version,
		Type:    common.MessageType(chunks[1]),
		Payload: chunks[2:],
	}, nil
}

func generateRawMessage(msg common.Message) string {
	return string(msg.Version) + common.Separator +
		string(msg.Type) + common.Separator +
		strings.Join(msg.Payload, common.Separator) + common.Terminator
}
