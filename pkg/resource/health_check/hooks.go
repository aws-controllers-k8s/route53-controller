package health_check

import (
	"fmt"
	"time"
)

// getCallerReference will generate a CallerReference for a given health check
// using the current timestamp, so that it produces a unique value
func getCallerReference() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}
