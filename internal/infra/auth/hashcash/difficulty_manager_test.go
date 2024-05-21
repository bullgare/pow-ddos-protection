package hashcash

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_DifficultyManager_calculateRPS(t *testing.T) {
	tt := []struct {
		name          string
		curReqBucket  int64
		prevReqBucket int64
		timeElapsed   time.Duration
		tickDuration  time.Duration
		expected      float64
	}{
		{
			name:          "happy path",
			curReqBucket:  50,
			prevReqBucket: 100,
			timeElapsed:   500 * time.Millisecond,
			tickDuration:  1000 * time.Millisecond,
			expected:      100,
		},
		{
			name:          "time elapsed is fixed when it's greater than tick duration",
			curReqBucket:  50,
			prevReqBucket: 100,
			timeElapsed:   1500 * time.Millisecond,
			tickDuration:  1000 * time.Millisecond,
			expected:      50,
		},
		{
			name:          "tiny tick durations are okay",
			curReqBucket:  50,
			prevReqBucket: 100,
			timeElapsed:   1 * time.Millisecond,
			tickDuration:  2 * time.Millisecond,
			expected:      50000,
		},
		{
			name:          "large tick durations are okay",
			curReqBucket:  50,
			prevReqBucket: 100,
			timeElapsed:   100 * time.Second,
			tickDuration:  200 * time.Second,
			expected:      0.5,
		},
		{
			name:          "tiny time elapsed is okay",
			curReqBucket:  50,
			prevReqBucket: 100,
			timeElapsed:   10 * time.Millisecond,
			tickDuration:  1000 * time.Millisecond,
			expected:      149,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := (&DifficultyManager{}).calculateRPS(tc.curReqBucket, tc.prevReqBucket, tc.timeElapsed, tc.tickDuration)

			assert.Equal(t, tc.expected, res)
		})
	}
}
