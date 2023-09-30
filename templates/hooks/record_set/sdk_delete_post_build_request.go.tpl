
	action := svcsdk.ChangeActionDelete
	recordSet, err := rm.newResourceRecordSet(ctx, r)
	if err != nil {
		return nil, err
	}
	changeBatch := rm.newChangeBatch(action, recordSet)
	input.SetChangeBatch(changeBatch)
