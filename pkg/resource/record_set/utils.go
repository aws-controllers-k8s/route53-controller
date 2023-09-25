package record_set

import (
	"strings"
)

// decodeRecordName filters the DNSName of a record set. ListResourceRecordSets returns
// the DNS name with a "." at the end and encodes the value for "*", so the DNSName needs
// to be filtered before comparing with our spec values.
func decodeRecordName(name string) string {
	if name[len(name)-1:] == "." {
		name = name[:len(name)-1]
	}
	if strings.Contains(name, "\\052") {
		return strings.Replace(name, "\\052", "*", -1)
	}
	return name
}
