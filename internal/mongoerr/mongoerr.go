package mongoerr

import (
	"context"
	"errors"
	"strings"
)

func IsTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection") ||
		strings.Contains(msg, "server selection") ||
		strings.Contains(msg, "timeout")
}
