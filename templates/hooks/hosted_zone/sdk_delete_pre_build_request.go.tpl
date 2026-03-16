	// Disassociate additional VPCs before deletion.
	// Route53 rejects DeleteHostedZone if >1 VPC is associated.
	if len(r.ko.Spec.AdditionalVPCs) > 0 {
		empty := rm.concreteResource(r.DeepCopy())
		empty.ko.Spec.AdditionalVPCs = nil
		if err = rm.syncVPCAssociations(ctx, rm.sdkapi, empty, r); err != nil {
			return nil, err
		}
	}
