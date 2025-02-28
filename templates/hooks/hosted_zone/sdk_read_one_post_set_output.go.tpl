	if err := rm.setResourceAdditionalFields(ctx, ko); err != nil {
		return nil, err
	}

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