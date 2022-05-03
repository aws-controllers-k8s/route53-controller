package hosted_zone

import (
	"fmt"
	"time"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
)

// getCallerReference will generate a CallerReference for a given hosted zone
// using the name of the zone and the current timestamp, so that it produces a
// unique value
func getCallerReference(zone *svcapitypes.HostedZone) string {
	return fmt.Sprintf("%s-%d", *zone.Spec.Name, time.Now().UnixMilli())
}
