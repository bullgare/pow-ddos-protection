package common

import (
	"context"
)

type Request struct {
	Type    MessageType
	Payload []string
}

type Response struct {
	Type    MessageType
	Payload []string
}

type HandlerFunc func(context.Context, Request) (Response, error)
