
	// Status represents whether record changes have been fully propagated to all
	// Route 53 authoritative DNS servers. The current status for the propagation
	// should be updated if it's not already INSYNC.
	err = rm.syncStatus(ctx, ko)
	if err != nil {
		return nil, err
	}
	if ko.Status.Status == nil || *ko.Status.Status != svcsdk.ChangeStatusInsync {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	}
