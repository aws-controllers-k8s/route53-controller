// Disassociate all but one VPC before deletion.
// Route53 rejects DeleteHostedZone if >1 VPC is associated.
//
// spec.vpcs path: keep Spec.VPCs[0] as final VPC, unset Spec.VPC to ensure Spec.VPCs and Spec.VPC 
// aren't set at the same time. 
// DeleteHostedZone removes the zone and its final VPC association.
if shouldRunVPCPreCleanup(r) {
	desired := rm.concreteResource(r.DeepCopy())
	if len(r.ko.Spec.VPCs) > 0 {
		desired.ko.Spec.VPC = nil
		desired.ko.Spec.VPCs = []*svcapitypes.VPC{r.ko.Spec.VPCs[0]}
	}
	if err = rm.syncVPCAssociations(ctx, rm.sdkapi, desired, r); err != nil {
		return nil, err
	}
}
