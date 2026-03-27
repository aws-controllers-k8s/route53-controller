// Disassociate all but one VPC before deletion.
// Route53 rejects DeleteHostedZone if >1 VPC is associated.
// Use Status.AssociatedVPCs (not Spec.VPCs) to check AWS actual state — a
// failed prior sync could leave more VPCs associated than the spec reflects.
//
// spec.vpcs path: keep Spec.VPCs[0] (user's intended primary).
// spec.vpc path: keep Spec.VPC (already in desired via DeepCopy, Spec.VPCs stays nil).
// DeleteHostedZone removes the zone and its final VPC association.
if shouldRunVPCPreCleanup(r) {
	desired := rm.concreteResource(r.DeepCopy())
	if len(r.ko.Spec.VPCs) > 0 {
		desired.ko.Spec.VPCs = []*svcapitypes.VPC{r.ko.Spec.VPCs[0]}
	}
	// spec.vpc path: desired.ko.Spec.VPC is already set from DeepCopy.
	// desired.ko.Spec.VPCs is nil, so syncVPCAssociations uses Spec.VPC as
	// the sole desired VPC and disassociates all others.
	if err = rm.syncVPCAssociations(ctx, rm.sdkapi, desired, r); err != nil {
		return nil, err
	}
}
