
	// Status represents whether record changes have been fully propagated to all
	// Route 53 authoritative DNS servers. The current status for the propagation
	// should be updated if it's not already INSYNC.
	//
	// If there is no change ID (e.g. the resource was adopted rather than
	// created/updated by this controller), we skip the propagation check.
	// The record was confirmed to exist via ListResourceRecordSets, so it is
	// already present on the authoritative servers.
	if ko.Status.ID != nil {
		err = rm.syncStatus(ctx, ko)
		if err != nil {
			return nil, err
		}
		if ko.Status.Status == nil || svcsdktypes.ChangeStatus(*ko.Status.Status) != svcsdktypes.ChangeStatusInsync {
			ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
		}
	}
