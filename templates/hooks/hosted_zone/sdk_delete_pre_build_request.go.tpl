// Disassociate all but one VPC before deletion.
// Route53 rejects DeleteHostedZone if >1 VPC is associated.
// Spec.VPCs is populated by sdkFind with the authoritative AWS state.
//
// Keep Spec.VPCs[0] as the final VPC. Null Spec.VPC to avoid the
// mutual-exclusivity guard in syncVPCAssociations.
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
