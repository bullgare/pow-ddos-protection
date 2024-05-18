package listener

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/users"
)

const (
	processConnTimeout = 500 * time.Millisecond
)

func New(
	address string,
	onError func(error),
	shareInfo func(string),
) (*Listener, error) {
	if address == "" {
		return nil, errors.New("address is required")
	}
	if onError == nil {
		return nil, errors.New("onError is required")
	}
	if shareInfo == nil {
		return nil, errors.New("shareInfo is required")
	}

	chQuit := make(chan struct{})

	return &Listener{
		address:   address,
		onError:   onError,
		shareInfo: shareInfo,
		chQuit:    chQuit,
	}, nil
}

type Listener struct {
	address   string
	onError   func(error)
	shareInfo func(string)
	chQuit    chan struct{}
}

func (l *Listener) StartWithHandlerFunc(ctx context.Context, handler common.HandlerFunc) error {
	if handler == nil {
		return errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-l.chQuit
		cancel()
	}()

	lsn, err := (&net.ListenConfig{}).Listen(ctx, "tcp", l.address)
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
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := lsn.Accept()
				if err != nil {
					l.onError(fmt.Errorf("accepting tcp connection: %w", err))
					continue
				}

				go l.processConnRequests(ctx, conn, handler)
			}
		}
	}()
}

func (l *Listener) processConnRequests(
	parentCtx context.Context,
	conn net.Conn,
	handler common.HandlerFunc,
) {
	defer func() { _ = conn.Close() }()

	// setting a timeout for connection to not exhaust number of available connections.
	_ = conn.SetDeadline(time.Now().Add(processConnTimeout))

	l.shareInfo(fmt.Sprintf("accepted connection from %s", conn.RemoteAddr()))
	defer func() {
		l.shareInfo(fmt.Sprintf("processed connection from %s", conn.RemoteAddr()))
	}()

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
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return
			}
			if err != nil {
				l.onError(fmt.Errorf("reading from conn: %w", err))
				continue
			}

			// not ideal, but we always restart the connection on the client side, that's why we cannot rely on a port.
			remoteAddress := conn.RemoteAddr().String()
			remoteAddressChunks := strings.Split(remoteAddress, ":")
			if len(remoteAddressChunks) > 1 {
				remoteAddress = strings.Join(remoteAddressChunks[:len(remoteAddressChunks)-1], ":")
			}
			ctx := users.NewContext(parentCtx, users.User{
				RemoteAddress: remoteAddress,
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
