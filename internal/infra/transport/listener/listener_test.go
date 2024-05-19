package listener_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol/common"
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
	validReq := common.Request{
		Type:    common.MessageTypeClientAuthReq,
		Payload: []string{""},
	}
	validResp := common.Response{
		Type:    common.MessageTypeSrvAuthResp,
		Payload: []string{"seed"},
	}

	tt := []struct {
		name          string
		req           common.Request
		resp          common.Response
		err           error
		expected      common.Response
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
			req:           common.Request{Type: "unknown req"},
			resp:          validResp,
			err:           nil,
			expected:      common.Response{Type: common.MessageTypeError, Payload: []string{"unexpected message type \"unknown req\""}},
			expectedError: assert.NoError,
		},
		{
			name:          "server error is returned properly",
			req:           validReq,
			resp:          validResp,
			err:           errors.New("some error"),
			expected:      common.Response{Type: common.MessageTypeError, Payload: []string{"some error"}},
			expectedError: assert.NoError,
		},
		{
			name:          "server returning a wrong message is sent as an error",
			req:           validReq,
			resp:          common.Response{Type: "unknown resp", Payload: nil},
			err:           nil,
			expected:      common.Response{},
			expectedError: assertion.ErrorWithMessage("parsing response: unexpected message type \"unknown resp\""),
		},
	}

	for _, tc := range tt {
		s.T().Run(tc.name, func(t *testing.T) {
			lsn, err := listener.New(validAddress, func(_ error) {}, func(_ string) {})
			require.NoError(t, err, "creating listener")
			defer lsn.Stop()
			go func() {
				err = lsn.StartWithHandlerFunc(context.Background(), func(_ context.Context, req common.Request) (common.Response, error) {
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

func (s *ListenerTestSuite) sendTCPRequest(t *testing.T, ctx context.Context, address string, req common.Request) (common.Response, error) {
	cl, err := client.New(address)
	require.NoError(t, err, "creating client")

	return cl.SendRequest(ctx, req)
}
