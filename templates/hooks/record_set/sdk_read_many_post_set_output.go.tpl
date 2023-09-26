
    // Status represents whether record changes have been fully propagated to all
    // Route 53 authoritative DNS servers. The current status for the propagation
    // should be updated per reconciliation
    if ko.Status.Status != nil {
        err = rm.syncStatus(ctx, ko)
        if err != nil {
            return nil, err
        }
    }
