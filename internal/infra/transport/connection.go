package transport

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol"
)

var MessageDelimiter = []byte(protocol.Terminator)[0]

func ParseRawMessage(raw string) (protocol.Message, error) {
	raw = strings.TrimSpace(raw)
	chunks := strings.Split(raw, protocol.Separator)
	if len(chunks) < 3 {
		return protocol.Message{}, fmt.Errorf("expected message to have at least 3 parts: version, message type and payload, got %d", len(chunks))
	}

	version := protocol.MessageVersion(chunks[0])
	if version != protocol.MessageVersionV1 {
		return protocol.Message{}, fmt.Errorf("unexpected message version %s, should be %s", version, protocol.MessageVersionV1)
	}

	messageType := chunks[1]
	switch protocol.MessageType(messageType) {
	case protocol.MessageTypeError,
		protocol.MessageTypeClientAuthReq,
		protocol.MessageTypeSrvAuthResp,
		protocol.MessageTypeClientDataReq,
		protocol.MessageTypeSrvDataResp:
	default:
		return protocol.Message{}, fmt.Errorf("unexpected message type %q", messageType)
	}

	return protocol.Message{
		Version: version,
		Type:    protocol.MessageType(chunks[1]),
		Payload: chunks[2:],
	}, nil
}

func SendMessage(w *bufio.Writer, msg protocol.Message) error {
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

func generateRawMessage(msg protocol.Message) string {
	return string(msg.Version) + protocol.Separator +
		string(msg.Type) + protocol.Separator +
		strings.Join(msg.Payload, protocol.Separator) + protocol.Terminator
}
