package common

import (
	"context"
	"time"
)

type RequestMeta struct {
	RemoteAddress string
	Time          time.Time
}

type Request struct {
	Type    MessageType
	Meta    RequestMeta
	Payload []string
}

type Response struct {
	Type    MessageType
	Payload []string
}

type HandlerFunc func(context.Context, Request) (Response, error)

type Handler interface {
	Handle(context.Context, Request) (Response, error)
}
