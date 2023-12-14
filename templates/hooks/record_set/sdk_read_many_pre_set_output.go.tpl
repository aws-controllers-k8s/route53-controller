
	// ListResourceRecordSets does not result in an exact match of relevant records as
	// it just consumes starting values for HostedZoneID, Name, RecordType, and SetIdentifier
	// from an alphabetically sorted list. As an example, if we are filtering for 'A' records,
	// ListResourceRecordSets could still return 'CNAME' records.
	var recordSets []*svcsdk.ResourceRecordSet
	for _, elem := range resp.ResourceRecordSets {
		if elem.Name != nil {
			// ListResourceRecordSets returns the full DNS name, so we need to reconstruct
			// the output to compare with the user specified subdomain. If a '*' value is
			// in the subdomain, ListResourceRecordSets returns it as an encoded value, so
			// this needs to be decoded before our comparison.
			subdomain := strings.TrimSuffix(*elem.Name, domain)
			subdomain = decodeRecordName(subdomain)

			// If user supplied no subdomain, we know that records with subdomains cannot
			// be a match and vice versa.
			if (r.ko.Spec.Name == nil && subdomain != "") || (r.ko.Spec.Name != nil && subdomain == "") {
				continue
			}

			// For cases where the user supplied a value to Spec.Name, irrelevant records
			// from ListResourceRecordSets will be further filtered out at a later point in
			// sdkFind. For now, parse out the "." at the end of the returned subdomain.
			if subdomain != "" {
				subdomain = subdomain[:len(subdomain)-1]
				elem.Name = &subdomain
			} else {
				elem.Name = nil
			}
		}

		// Similar to above, remove the "." at the end and decode the "*" value as necessary.
		if elem.AliasTarget != nil && ko.Spec.AliasTarget != nil {
			if elem.AliasTarget.DNSName != nil && ko.Spec.AliasTarget.DNSName != nil {
				dnsName = *elem.AliasTarget.DNSName
				decodedName := decodeRecordName(dnsName[:len(dnsName)-1])
				elem.AliasTarget.DNSName = &decodedName
			}
		}

		// RecordTypes are required, so discard records that don't have them.
		if elem.Type == nil || (*elem.Type != *ko.Spec.RecordType) {
			continue
		}

		recordSets = append(recordSets, elem)
	}
	resp.ResourceRecordSets = recordSets
