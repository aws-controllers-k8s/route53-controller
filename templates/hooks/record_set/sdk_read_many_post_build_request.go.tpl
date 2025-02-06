
	// Retrieve the domain name of the hosted zone through the ID.
	domain, err := rm.getHostedZoneDomain(ctx, r)
	if err != nil {
		return nil, err
	}

	// Return the combined value of the user specified subdomain and the hosted zone domain.
	dnsName := rm.getDNSName(r, domain)

	// Setting the starting point to the following values reduces the number of irrelevant
	// records that are returned.
	input.StartRecordName = &dnsName
	if r.ko.Spec.RecordType != nil {
		input.StartRecordType = svcsdktypes.RRType(*r.ko.Spec.RecordType)
	}
	if r.ko.Spec.SetIdentifier != nil {
		input.StartRecordIdentifier = r.ko.Spec.SetIdentifier
	}
