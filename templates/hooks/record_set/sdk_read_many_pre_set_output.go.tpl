
	// ListResourceRecordSets does not result in an exact match even after specifying
	// all of hostedZoneID, Name, RecordType, and SetIdentifier. While all filtering
	// can be done with list_operation.match_fields, the below allows "json:recordType"
	// to be used instead of "json:type_".
	var recordSets []*svcsdk.ResourceRecordSet
	for _, elem := range resp.ResourceRecordSets {
		if elem.Name != nil {
			decodedName := decodeRecordName(*elem.Name, *ko.Spec.Name)
			elem.Name = &decodedName
		}

		if elem.AliasTarget != nil && ko.Spec.AliasTarget != nil {
			if elem.AliasTarget.DNSName != nil && ko.Spec.AliasTarget.DNSName != nil {
				decodedName := decodeRecordName(*elem.AliasTarget.DNSName, *ko.Spec.AliasTarget.DNSName)
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
