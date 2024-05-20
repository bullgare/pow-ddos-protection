package transport

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bullgare/pow-ddos-protection/internal/infra/protocol"
	"github.com/bullgare/pow-ddos-protection/pkg/assertion"
)

func Test_ParseRawMessage(t *testing.T) {
	tt := []struct {
		name          string
		raw           string
		expected      protocol.Message
		expectedError assert.ErrorAssertionFunc
	}{
		{
			name: "happy path",
			raw:  "v1|c2s_data_req|token|seed",
			expected: protocol.Message{
				Version: protocol.MessageVersionV1,
				Type:    protocol.MessageTypeClientDataReq,
				Payload: []string{"token", "seed"},
			},
			expectedError: assert.NoError,
		},
		{
			name:          "not enough chunks - error",
			raw:           "c2s_data_req|seed",
			expected:      protocol.Message{},
			expectedError: assertion.ErrorWithMessage("expected message to have at least 3 parts: version, message type and payload, got 2"),
		},
		{
			name:          "unexpected version - error",
			raw:           "v2|c2s_data_req|token|seed",
			expected:      protocol.Message{},
			expectedError: assertion.ErrorWithMessage("unexpected message version v2, should be v1"),
		},
		{
			name:          "unexpected message type - error",
			raw:           "v1|unknown_type|token|seed",
			expected:      protocol.Message{},
			expectedError: assertion.ErrorWithMessage("unexpected message type \"unknown_type\""),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := ParseRawMessage(tc.raw)

			require.True(t, tc.expectedError(t, err))
			assert.Equal(t, tc.expected, res)
		})
	}
}

func Test_SendMessage(t *testing.T) {
	tt := []struct {
		name          string
		msg           protocol.Message
		expected      string
		expectedError assert.ErrorAssertionFunc
	}{
		{
			name: "happy path",
			msg: protocol.Message{
				Version: protocol.MessageVersionV1,
				Type:    protocol.MessageTypeClientDataReq,
				Payload: []string{"token", "seed"},
			},
			expected:      "v1|c2s_data_req|token|seed\n",
			expectedError: assert.NoError,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var res bytes.Buffer
			w := bufio.NewWriter(&res)

			err := SendMessage(w, tc.msg)

			require.True(t, tc.expectedError(t, err))
			assert.Equal(t, tc.expected, res.String())
		})
	}
}

func Test_generateRawMessage(t *testing.T) {
	tt := []struct {
		name     string
		msg      protocol.Message
		expected string
	}{
		{
			name: "happy path",
			msg: protocol.Message{
				Version: protocol.MessageVersionV1,
				Type:    protocol.MessageTypeSrvAuthResp,
				Payload: []string{"1", "2", "3"},
			},
			expected: "v1|s2c_auth_resp|1|2|3\n",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := generateRawMessage(tc.msg)

			assert.Equal(t, tc.expected, res)
		})
	}
}
