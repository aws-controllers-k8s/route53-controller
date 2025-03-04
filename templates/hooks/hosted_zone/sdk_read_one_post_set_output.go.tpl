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
