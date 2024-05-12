package listener

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
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/users"
)

func New(
	address string,
	onError func(error),
) (*Listener, error) {
	if address == "" {
		return nil, errors.New("address is required")
	}
	if onError == nil {
		return nil, errors.New("onError is required")
	}

	chQuit := make(chan struct{})

	return &Listener{
		address: address,
		onError: onError,
		chQuit:  chQuit,
	}, nil
}

type Listener struct {
	address string
	onError func(error)
	chQuit  chan struct{}
}

func (l *Listener) StartWithHandlerFunc(ctx context.Context, handler common.HandlerFunc) error {
	if handler == nil {
		return errors.New("handler is required")
	}

	lsn, err := net.Listen("tcp", l.address)
	if err != nil {
		return fmt.Errorf("listening %s: %w", l.address, err)
	}
	defer func() { _ = lsn.Close() }()

	l.handleConnections(ctx, lsn, handler)
	<-l.chQuit

	return nil
}

func (l *Listener) Stop() {
	close(l.chQuit)
}

func (l *Listener) handleConnections(ctx context.Context, lsn net.Listener, handler common.HandlerFunc) {
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				close(l.chQuit)
				return
			case <-l.chQuit:
				return
			default:
				conn, err := lsn.Accept()
				if err != nil {
					l.onError(fmt.Errorf("accepting tcp connection: %w", err))
					continue
				}

				go l.processRequests(ctx, conn, handler)
			}
		}
	}()
}

// processRequests handles all requests for one connection.
// TODO we might limit connection time to not exhaust the connections.
func (l *Listener) processRequests(
	parentCtx context.Context,
	conn net.Conn,
	handler common.HandlerFunc,
) {
	defer func() { _ = conn.Close() }()

	if handler == nil {
		l.onError(errors.New("handler is required"))
		return
	}

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	for {
		select {
		case <-parentCtx.Done():
			return
		default:
			raw, err := r.ReadString(transport.MessageDelimiter)
			if err == io.EOF {
				return
			}
			if err != nil {
				l.onError(fmt.Errorf("reading from conn: %w", err))
				continue
			}
			ctx := users.NewContext(parentCtx, users.User{
				RemoteAddress: conn.RemoteAddr().String(),
				RequestTime:   time.Now(),
			})

			msg, err := transport.ParseRawMessage(raw)
			if err != nil {
				l.onError(fmt.Errorf("parsing request: %w", err))
				l.sendResponse(w, common.Response{Type: common.MessageTypeError, Payload: []string{err.Error()}})
				continue
			}

			req := common.Request{
				Type:    msg.Type,
				Payload: slices.Clone(msg.Payload),
			}

			resp, err := handler(ctx, req)
			if err != nil {
				l.onError(err)
				l.sendResponse(w, common.Response{Type: common.MessageTypeError, Payload: []string{err.Error()}})
				continue
			}

			l.sendResponse(w, resp)
		}
	}
}

func (l *Listener) sendResponse(w *bufio.Writer, resp common.Response) {
	msg := common.Message{
		Version: common.MessageVersionV1,
		Type:    resp.Type,
		Payload: slices.Clone(resp.Payload),
	}

	err := transport.SendMessage(w, msg)
	if err != nil {
		l.onError(err)
	}
}
