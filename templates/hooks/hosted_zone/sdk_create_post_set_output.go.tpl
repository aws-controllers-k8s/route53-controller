	if ko.Status.ID != nil {
		latest := &resource{}
		latest.ko = &svcapitypes.HostedZone{}
		latest.ko.Status.ID = ko.Status.ID

		if resp.DelegationSet != nil {
			f := &svcapitypes.DelegationSet{}
			if resp.DelegationSet.CallerReference != nil {
				f.CallerReference = resp.DelegationSet.CallerReference
			}
			if resp.DelegationSet.Id != nil {
				f.ID = resp.DelegationSet.Id
			}
			if resp.DelegationSet.NameServers != nil {
				f.NameServers = aws.StringSlice(resp.DelegationSet.NameServers)
			}
			ko.Status.DelegationSet = f
		} else {
			ko.Status.DelegationSet = nil
		}

		// This is create operation. So, no tags are present in HostedZone.
		// So, 'latest' is empty except we have copied 'ID' into the status to
		// make syncTags() happy.
		if err := rm.syncTags(ctx, desired, latest); err != nil {
			return nil, err
		}
	}
