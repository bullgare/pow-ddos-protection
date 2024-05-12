package connection

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
)

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

// Connection knows how to send and receive messages over a
type Connection struct {
	onError   func(error)
	delimiter byte
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

func (c *Connection) parseRawMessage(raw string) (common.Message, error) {
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
