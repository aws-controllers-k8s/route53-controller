
	// Retrieve the domain name of the hosted zone through the ID.
	domain, err := rm.getHostedZoneDomain(ctx, r)
	if err != nil {
		return nil, err
	}

	// Return the combined value of the user specified subdomain and the hosted zone domain.
	dnsName := rm.getDNSName(r, domain)

	// Setting the starting point to the following values reduces the number of irrelevant
	// records that are returned.
	input.SetStartRecordName(dnsName)
	if r.ko.Spec.RecordType != nil {
		input.SetStartRecordType(*r.ko.Spec.RecordType)
	}
	if r.ko.Spec.SetIdentifier != nil {
		input.SetStartRecordIdentifier(*r.ko.Spec.SetIdentifier)
	}
