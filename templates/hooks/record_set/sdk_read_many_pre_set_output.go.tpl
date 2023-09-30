
	// ListResourceRecordSets does not result in an exact match even after specifying
	// all of hostedZoneID, Name, RecordType, and SetIdentifier. While all filtering
	// can be done with list_operation.match_fields, the below allows "json:recordType"
	// to be used instead of "json:type_".
	var recordSets []*svcsdk.ResourceRecordSet
	for _, elem := range resp.ResourceRecordSets {

		if elem.Name != nil {
			// Take the subdomain value from the returned results to compare them with
			// the user specified subdomain
			subdomain := strings.TrimSuffix(*elem.Name, domain)
			subdomain = decodeRecordName(subdomain)

			// If user supplied no subdomain, we know that records with subdomains cannot
			// be a match
			if r.ko.Spec.Name == nil && subdomain != "" {
				continue
			}

			if r.ko.Spec.Name == nil && subdomain == "" {
				elem.Name = nil
			}

			// For cases where the user did supply a subdomain value, irrelevant records
			// returned from ListResourceRecordSets will be further filtered out at a
			// later point in this method. For now, just parse out the "." at the end
			// of the decoded subdomain
			if subdomain != "" {
				subdomain = subdomain[:len(subdomain)-1]
				elem.Name = &subdomain
			}
		}

		if elem.AliasTarget != nil && ko.Spec.AliasTarget != nil {
			if elem.AliasTarget.DNSName != nil && ko.Spec.AliasTarget.DNSName != nil {
				filteredName := filterRecordName(*elem.AliasTarget.DNSName, *ko.Spec.AliasTarget.DNSName)
				decodedName := decodeRecordName(filteredName)
				elem.AliasTarget.DNSName = &decodedName
			}
		}

		// RecordTypes are required, so discard records that don't have them
		if elem.Type == nil || (*elem.Type != *ko.Spec.RecordType) {
			continue
		}

		recordSets = append(recordSets, elem)
	}
	resp.ResourceRecordSets = recordSets
