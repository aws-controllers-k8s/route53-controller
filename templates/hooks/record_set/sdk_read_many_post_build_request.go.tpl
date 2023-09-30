
	// Setting the starting point to these values reduces the number of total records
	// that are returned
	domain, err := rm.getHostedZoneDomain(ctx, r)
	if err != nil {
		return nil, err
	}

	dnsName := rm.getDNSName(r, domain)
	input.SetStartRecordName(dnsName)

	if r.ko.Spec.RecordType != nil {
		input.SetStartRecordType(*r.ko.Spec.RecordType)
	}
	if r.ko.Spec.SetIdentifier != nil {
		input.SetStartRecordIdentifier(*r.ko.Spec.SetIdentifier)
	}
