package common

import (
	"fmt"
)

type MessageVersion string

const MessageVersionV1 MessageVersion = "v1"

type MessageType string

const (
	MessageTypeError         MessageType = "error"
	MessageTypeClientAuthReq MessageType = "c2s_auth_req"
	MessageTypeSrvAuthResp   MessageType = "s2c_auth_resp"
	MessageTypeClientDataReq MessageType = "c2s_data_req"
	MessageTypeSrvDataResp   MessageType = "s2c_data_resp"
)

const (
	Separator  = "|"
	Terminator = "\n" // should be one byte to not break connection reader
)

type Message struct {
	Version MessageVersion
	Type    MessageType
	Payload []string
}

func MapPayloadToTokenAndSeed(payload []string) (token, seed string, err error) {
	if len(payload) != 2 {
		return "", "", fmt.Errorf("expected a payload to consist of 2 parts, got %d", len(payload))
	}

	return payload[0], payload[1], nil
}

func GeneratePayloadFromTokenAndSeed(token, seed string) []string {
	return []string{token, seed}
}
