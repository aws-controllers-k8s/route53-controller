// Disassociate all but vpcs[0] before deletion.
// Route53 rejects DeleteHostedZone if >1 VPC is associated.
// Use Status.AssociatedVPCs (not Spec.VPCs) to check AWS actual state — a
// failed prior sync could leave more VPCs associated than the spec reflects.
// Keep Spec.VPCs[0] (the user's intended primary) and disassociate the rest.
// DeleteHostedZone removes the zone and its final VPC association.
//
// Note: users on the legacy spec.vpc path who have manually associated additional
// VPCs out-of-band must disassociate them manually before deletion — this
// controller does not clean them up on the spec.vpc path.
if shouldRunVPCPreCleanup(r) {
	desired := rm.concreteResource(r.DeepCopy())
	desired.ko.Spec.VPCs = []*svcapitypes.VPC{r.ko.Spec.VPCs[0]}
	if err = rm.syncVPCAssociations(ctx, rm.sdkapi, desired, r); err != nil {
		return nil, err
	}
}
