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

	// Populate Spec.VPCs from the authoritative AWS VPC list so that
	// compareVPCs / syncVPCAssociations can compare desired vs actual.
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
	// Validate spec.name matches the actual hosted zone name returned by AWS.
	// Zone names are immutable in Route53 — a mismatch during adoption means
	// the user pointed the wrong zone ID at this CR. Return a terminal error
	// so the mismatch is visible in status.conditions rather than silently ignored.
	if r.ko.Spec.Name != nil && ko.Spec.Name != nil {
		if *r.ko.Spec.Name != *ko.Spec.Name {
			return nil, ackerr.NewTerminalError(fmt.Errorf(
				"spec.name %q does not match hosted zone name %q: "+
					"correct spec.name to match the actual zone name",
				*r.ko.Spec.Name, *ko.Spec.Name,
			))
		}
	}