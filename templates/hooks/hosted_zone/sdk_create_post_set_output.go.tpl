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

		// Seed the current VPC list from the create response so that
		// syncVPCAssociations knows the initial VPC is already associated
		// and does not attempt to re-associate it.
		if resp.VPC != nil && resp.VPC.VPCId != nil {
			region := string(resp.VPC.VPCRegion)
			latest.ko.Status.AssociatedVPCs = []*svcapitypes.VPC{
				{
					VPCID:     resp.VPC.VPCId,
					VPCRegion: &region,
				},
			}
		}

		// This is create operation. So, no tags are present in HostedZone.
		// So, 'latest' is empty except we have copied 'ID' into the status to
		// make syncTags() happy.
		if err := rm.syncTags(ctx, desired, latest); err != nil {
			return nil, err
		}

		// Sync additional VPC associations during create. The new zone already
		// has resp.VPC associated; syncVPCAssociations will associate the rest.
		// Return ko (not nil) on error so status.id is written to k8s.
		if err := rm.syncVPCAssociations(ctx, rm.sdkapi, desired, latest); err != nil {
			return &resource{ko}, err
		}
	}