package listener_test

import (
	"context"
	"errors"
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

type ListenerTestSuite struct {
	suite.Suite
}

func Test_ListenerTestSuite(t *testing.T) {
	suite.Run(t, new(ListenerTestSuite))
}

// tests are a bit fragile as they involve network communication, more time needs to be invested here.
func (s *ListenerTestSuite) Test_StartWithHandlerFunc() {
	validAddress := "127.0.0.1:33411"
	validReq := protocol.Request{
		Type:    protocol.MessageTypeClientAuthReq,
		Payload: []string{""},
	}
	validResp := protocol.Response{
		Type:    protocol.MessageTypeSrvAuthResp,
		Payload: []string{"seed"},
	}

	tt := []struct {
		name          string
		req           protocol.Request
		resp          protocol.Response
		err           error
		expected      protocol.Response
		expectedError assert.ErrorAssertionFunc
	}{
		{
			name:          "happy path",
			req:           validReq,
			resp:          validResp,
			err:           nil,
			expected:      validResp,
			expectedError: assert.NoError,
		},
		{
			name:          "invalid request - error",
			req:           protocol.Request{Type: "unknown req"},
			resp:          validResp,
			err:           nil,
			expected:      protocol.Response{Type: protocol.MessageTypeError, Payload: []string{"unexpected message type \"unknown req\""}},
			expectedError: assert.NoError,
		},
		{
			name:          "server error is returned properly",
			req:           validReq,
			resp:          validResp,
			err:           errors.New("some error"),
			expected:      protocol.Response{Type: protocol.MessageTypeError, Payload: []string{"some error"}},
			expectedError: assert.NoError,
		},
		{
			name:          "server returning a wrong message is sent as an error",
			req:           validReq,
			resp:          protocol.Response{Type: "unknown resp", Payload: nil},
			err:           nil,
			expected:      protocol.Response{},
			expectedError: assertion.ErrorWithMessage("parsing response: unexpected message type \"unknown resp\""),
		},
	}

	for _, tc := range tt {
		s.T().Run(tc.name, func(t *testing.T) {
			lsn, err := listener.New(validAddress, func(_ error) {}, func(_ string) {})
			require.NoError(t, err, "creating listener")
			defer lsn.Stop()
			go func() {
				err = lsn.StartWithHandlerFunc(context.Background(), func(_ context.Context, req protocol.Request) (protocol.Response, error) {
					require.Equal(t, tc.req, req)
					time.Sleep(20 * time.Millisecond)
					return tc.resp, tc.err
				})
			}()
			require.NoError(t, err, "lsn.StartWithHandlerFunc")
			time.Sleep(100 * time.Millisecond) // giving time to start the server

			resp, err := s.sendTCPRequest(t, context.Background(), validAddress, tc.req)

			require.True(t, tc.expectedError(t, err))
			assert.Equal(t, tc.expected, resp)
		})
	}
}

func (s *ListenerTestSuite) sendTCPRequest(t *testing.T, ctx context.Context, address string, req protocol.Request) (protocol.Response, error) {
	cl, err := client.New(address)
	require.NoError(t, err, "creating client")

	return cl.SendRequest(ctx, req)
}
