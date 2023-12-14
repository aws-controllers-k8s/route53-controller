
	action := svcsdk.ChangeActionCreate
	recordSet, err := rm.newResourceRecordSet(ctx, desired)
	if err != nil {
		return nil, err
	}
	changeBatch := rm.newChangeBatch(action, recordSet)
	input.SetChangeBatch(changeBatch)
