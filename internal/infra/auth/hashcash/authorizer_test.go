package hashcash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ucontracts "github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
	"github.com/bullgare/pow-ddos-protection/pkg/assertion"
)

func Test_Authorizer_MergeWithConfig(t *testing.T) {
	tt := []struct {
		name     string
		data     string
		cfg      ucontracts.AuthorizerConfig
		expected string
	}{
		{
			name:     "happy path",
			data:     "data",
			cfg:      ucontracts.AuthorizerConfig{DifficultyLevelPercent: 15},
			expected: "v1;15;data",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := (Authorizer{}).MergeWithConfig(tc.data, tc.cfg)

			assert.Equal(t, tc.expected, res)
		})
	}
}

func Test_Authorizer_ParseConfigFrom(t *testing.T) {
	tt := []struct {
		name           string
		dataWithConfig string
		expectedData   string
		expectedConfig ucontracts.AuthorizerConfig
		expectedError  assert.ErrorAssertionFunc
	}{
		{
			name:           "happy path",
			dataWithConfig: "v1;15;data",
			expectedData:   "data",
			expectedConfig: ucontracts.AuthorizerConfig{DifficultyLevelPercent: 15},
			expectedError:  assert.NoError,
		},
		{
			name:           "not enough chunks - error",
			dataWithConfig: "15;data",
			expectedData:   "",
			expectedConfig: ucontracts.AuthorizerConfig{},
			expectedError:  assertion.ErrorWithMessage("got 2 chunks in marshalled data with config instead of 3"),
		},
		{
			name:           "unexpected version - error",
			dataWithConfig: "v2;15;data",
			expectedData:   "",
			expectedConfig: ucontracts.AuthorizerConfig{},
			expectedError:  assertion.ErrorWithMessage("auth data version v2 is not supported"),
		},
		{
			name:           "config level is not an int - error",
			dataWithConfig: "v1;fifteen;data",
			expectedData:   "",
			expectedConfig: ucontracts.AuthorizerConfig{},
			expectedError:  assertion.ErrorWithMessage("parsing difficulty level percent: strconv.Atoi: parsing \"fifteen\": invalid syntax"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			data, cfg, err := (Authorizer{}).ParseConfigFrom(tc.dataWithConfig)

			require.True(t, tc.expectedError(t, err))
			assert.Equal(t, tc.expectedData, data)
			assert.Equal(t, tc.expectedConfig, cfg)
		})
	}
}

func Test_Authorizer_calculateBitsLen(t *testing.T) {
	tt := []struct {
		name       string
		bitsLenMin uint
		bitsLenMax uint
		cfg        ucontracts.AuthorizerConfig
		expected   uint
	}{
		{
			name:       "happy path",
			bitsLenMin: 0,
			bitsLenMax: 100,
			cfg:        ucontracts.AuthorizerConfig{DifficultyLevelPercent: 5},
			expected:   5,
		},
		{
			name:       "greater than max - fix",
			bitsLenMin: 2,
			bitsLenMax: 10,
			cfg:        ucontracts.AuthorizerConfig{DifficultyLevelPercent: 150},
			expected:   10,
		},
		{
			name:       "lower than min - fix",
			bitsLenMin: 2,
			bitsLenMax: 10,
			cfg:        ucontracts.AuthorizerConfig{DifficultyLevelPercent: -150},
			expected:   2,
		},
		{
			name:       "more complicated calculation",
			bitsLenMin: 20,
			bitsLenMax: 105,
			cfg:        ucontracts.AuthorizerConfig{DifficultyLevelPercent: 15},
			expected:   32,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := (Authorizer{bitsLenMin: tc.bitsLenMin, bitsLenMax: tc.bitsLenMax}).calculateBitsLen(tc.cfg)

			assert.Equal(t, tc.expected, res)
		})
	}
}
