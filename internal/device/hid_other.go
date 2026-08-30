//go:build !windows

package device

import "context"

func HasExpectedHID(context.Context) (bool, error) {
	return false, nil
}
