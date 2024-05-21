package client_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/client"
	"github.com/bullgare/pow-ddos-protection/internal/infra/transport/listener"
	"github.com/bullgare/pow-ddos-protection/pkg/assertion"
)

type ClientTestSuite struct {
	suite.Suite
}

func Test_ClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

// tests are a bit fragile as they involve network communication, more time needs to be invested here.
func (s *ClientTestSuite) Test_SendRequest() {
	validPort := 55144
	validReq := protocol.Request{
		Type:    protocol.MessageTypeClientAuthReq,
		Payload: []string{""},
	}
	validResp := protocol.Response{
		Type:    protocol.MessageTypeSrvAuthResp,
		Payload: []string{"seed"},
	}

	tt := []struct {
		name                string
		mockedServerHandler func(t *testing.T) protocol.HandlerFunc
		req                 protocol.Request
		expected            protocol.Response
		expectedError       assert.ErrorAssertionFunc
	}{
		{
			name: "happy path",
			mockedServerHandler: func(t *testing.T) protocol.HandlerFunc {
				return func(_ context.Context, req protocol.Request) (protocol.Response, error) {
					assert.Equal(t, validReq, req)
					return validResp, nil
				}
			},
			req:           validReq,
			expected:      validResp,
			expectedError: assert.NoError,
		},
		{
			name: "connection timeout -> error",
			mockedServerHandler: func(t *testing.T) protocol.HandlerFunc {
				return func(_ context.Context, req protocol.Request) (protocol.Response, error) {
					// client's connectionTimeout == 500ms
					time.Sleep(700 * time.Millisecond)
					assert.Equal(t, validReq, req)
					return validResp, nil
				}
			},
			req:      validReq,
			expected: protocol.Response{},
			// this is a bit fragile as it depends on the tests run order
			expectedError: assertion.ErrorWithMessageContainsAny([]string{
				"connecting to \"127.0.0.1:55146\": dial tcp 127.0.0.1:55146: connect: connection refused",
				"->127.0.0.1:55146: i/o timeout",
			}),
		},
		{
			name: "unexpected response - error",
			mockedServerHandler: func(t *testing.T) protocol.HandlerFunc {
				return func(_ context.Context, req protocol.Request) (protocol.Response, error) {
					assert.Equal(t, validReq, req)
					return protocol.Response{}, nil
				}
			},
			req:           validReq,
			expected:      protocol.Response{},
			expectedError: assertion.ErrorWithMessage("parsing response: unexpected message type \"\""),
		},
	}

	for _, tc := range tt {
		s.T().Run(tc.name, func(t *testing.T) {
			validPort++
			address := "127.0.0.1:" + strconv.Itoa(validPort)
			srvStop := s.createTCPListener(t, address, tc.mockedServerHandler(t))
			defer srvStop()
			cl, err := client.New(address)
			require.NoError(t, err, "creating a client")

			resp, err := (cl).SendRequest(context.Background(), tc.req)

			require.True(t, tc.expectedError(t, err))
			assert.Equal(t, tc.expected, resp)
		})
	}
}

func (s *ClientTestSuite) createTCPListener(t *testing.T, address string, handler protocol.HandlerFunc) func() {
	ctx, cancel := context.WithCancel(context.Background())

	lsn, err := listener.New(address, func(_ error) {}, func(_ string) {})
	require.NoError(t, err, "creating listener")

	go func() {
		err = lsn.StartWithHandlerFunc(ctx, handler)
		require.NoError(t, err, "lsn.StartWithHandlerFunc")
	}()

	time.Sleep(100 * time.Millisecond) // giving time to start the server

	return func() {
		lsn.Stop()
		cancel()
	}
}
