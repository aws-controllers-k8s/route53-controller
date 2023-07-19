	if ko.Status.ID != nil {
		latest := &resource{}
		latest.ko = &svcapitypes.HostedZone{}
		latest.ko.Status.ID = ko.Status.ID

		// This is create operation. So, no tags are present in HostedZone.
		// So, 'latest' is empty except we have copied 'ID' into the status to
		// make syncTags() happy.
		if err := rm.syncTags(ctx, desired, latest); err != nil {
			return nil, err
		}
	}
