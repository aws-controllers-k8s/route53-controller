	if ko.Status.ID != nil {
		latest := &resource{}
		latest.ko = &svcapitypes.HostedZone{}
		latest.ko.Status.ID = ko.Status.ID

		if resp.DelegationSet != nil {
			f := &svcapitypes.DelegationSet{}
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

		// Sync additional VPC associations during create. The new zone has no
		// VPCs associated yet, so syncVPCAssociations will associate all desired
		// VPCs. Return ko (not nil) on error so status.id is written to k8s.
		if err := rm.syncVPCAssociations(ctx, rm.sdkapi, desired, latest); err != nil {
			return &resource{ko}, err
		}
	}
