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

		// If there are additional VPCs beyond the first (already associated by
		// create), requeue so the update path handles the remaining associations.
		// This ensures association errors are attributed to the sync phase, not
		// the create phase, and that status.id is always written to k8s on create.
		if len(desired.ko.Spec.VPCs) > 1 {
			ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse,
				aws.String("requeuing to associate additional VPCs"), nil)
			return &resource{ko}, ackrequeue.NeededAfter(
				fmt.Errorf("reconciling additional VPC associations"), 1*time.Second)
		}
	}