	if desired.ko.Spec.ChangeBatch != nil {
		rlog.Info("WARNING: ChangeBatch field is no-op, and will be removed soon!")
	}
	action := svcsdktypes.ChangeActionCreate
	recordSet, err := rm.newResourceRecordSet(ctx, desired)
	if err != nil {
		return nil, err
	}
	changeBatch := rm.newChangeBatch(action, recordSet)
	input.ChangeBatch = changeBatch
