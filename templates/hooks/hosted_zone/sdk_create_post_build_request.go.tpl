// Validate VPC field configuration before attempting create.
if err := validateVPCFields(desired.ko.Spec); err != nil {
	return nil, err
}

// If spec.vpcs is set (and spec.vpc is nil), inject vpcs[0] as the creation
// VPC. The generated newCreateRequestPayload leaves input.VPC nil when
// spec.vpc is nil. The user controls which VPC is primary by placing it first.
if len(desired.ko.Spec.VPCs) > 0 {
	input.VPC = &svcsdktypes.VPC{
		VPCId:     desired.ko.Spec.VPCs[0].VPCID,
		VPCRegion: svcsdktypes.VPCRegion(*desired.ko.Spec.VPCs[0].VPCRegion),
	}
}

// You must use a unique CallerReference string every time you submit a
// CreateHostedZone request. CallerReference can be any unique string, for
// example, a date/timestamp.
input.CallerReference = aws.String(getCallerReference(desired.ko))
