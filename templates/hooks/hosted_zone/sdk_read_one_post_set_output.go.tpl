	if err := rm.setResourceAdditionalFields(ctx, ko); err != nil {
		return nil, err
	}

	if resp.DelegationSet != nil {
		f := &svcapitypes.DelegationSet{}
		if resp.DelegationSet.NameServers != nil {
			f.NameServers = aws.StringSlice(resp.DelegationSet.NameServers)
		}
		ko.Status.DelegationSet = f
	} else {
		ko.Status.DelegationSet = nil
	}

	// Always record the authoritative VPC list in status so that
	// syncVPCAssociations can use it without an extra GetHostedZone call.
	ko.Spec.VPCs = nil
	for _, v := range resp.VPCs {
		if v.VPCId == nil {
			continue
		}
		region := string(v.VPCRegion)
		ko.Spec.VPCs = append(ko.Spec.VPCs, &svcapitypes.VPC{
			VPCID:     v.VPCId,
			VPCRegion: &region,
		})
	}