package wordofwisdom

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/client"
	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

func New(transClient *client.Client) (*Client, error) {
	if transClient == nil {
		return nil, errors.New("transport client is required")
	}

	return &Client{
		transClient: transClient,
	}, nil
}

var _ contracts.ClientWordOfWisdom = &Client{}

type Client struct {
	transClient *client.Client
}

func (c *Client) GetAuthParams(ctx context.Context) (contracts.AuthResponse, error) {
	request := common.Request{
		Type: common.MessageTypeClientAuthReq,
	}
	response, err := c.transClient.SendRequest(ctx, request)
	if err != nil {
		return contracts.AuthResponse{}, err
	}

	switch response.Type {
	case common.MessageTypeSrvAuthResp:
		return contracts.AuthResponse{Seed: response.Payload[0]}, nil
	case common.MessageTypeError:
		return contracts.AuthResponse{}, fmt.Errorf("server returned an error: %s", strings.Join(response.Payload, ";"))
	default:
		return contracts.AuthResponse{}, fmt.Errorf("server returned an unexpected message: %s instead of %s", response.Type, common.MessageTypeSrvAuthResp)
	}
}

func (c *Client) GetData(ctx context.Context, req contracts.DataRequest) (contracts.DataResponse, error) {
	request := common.Request{
		Type:    common.MessageTypeClientDataReq,
		Payload: common.GeneratePayloadFromTokenAndSeed(req.Token, req.OriginalSeed),
	}
	response, err := c.transClient.SendRequest(ctx, request)
	if err != nil {
		return contracts.DataResponse{}, err
	}

	switch response.Type {
	case common.MessageTypeSrvDataResp:
		return contracts.DataResponse{MyPrecious: response.Payload[0]}, nil
	case common.MessageTypeError:
		return contracts.DataResponse{}, fmt.Errorf("server returned an error: %s", strings.Join(response.Payload, ";"))
	default:
		return contracts.DataResponse{}, fmt.Errorf("server returned an unexpected message: %s instead of %s", response.Type, common.MessageTypeSrvDataResp)
	}
}
