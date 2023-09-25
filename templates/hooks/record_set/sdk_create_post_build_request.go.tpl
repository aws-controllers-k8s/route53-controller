
	action := svcsdk.ChangeActionCreate
	recordSet := rm.newResourceRecordSet(desired)
	changeBatch := rm.newChangeBatch(action, recordSet)
	input.SetChangeBatch(changeBatch)