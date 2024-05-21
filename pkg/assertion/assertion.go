package assertion

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/assert"
)

func ErrorWithMessage(msg string) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, resultErr error, msgAndArgs ...interface{}) bool {
		if resultErr == nil {
			return assert.Fail(t, "An error is expected but got nil.", msgAndArgs...)
		}

		return assert.EqualError(t, resultErr, msg, msgAndArgs...)
	}
}

func ErrorWithMessageContainsAny(msgs []string) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, resultErr error, msgAndArgs ...interface{}) bool {
		if resultErr == nil {
			return assert.Fail(t, "An error is expected but got nil.", msgAndArgs...)
		}

		if !assert.Error(t, resultErr, msgAndArgs...) {
			return false
		}

		actual := resultErr.Error()

		for _, msg := range msgs {
			if strings.Contains(actual, msg) {
				return true
			}
		}

		assert.Fail(t, fmt.Sprintf("Error %#v does not contain any of %#v", actual, msgs), msgAndArgs...)
		return false
	}
}
