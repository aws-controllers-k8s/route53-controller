
	action := svcsdk.ChangeActionDelete
	recordSet := rm.newResourceRecordSet(r)
	changeBatch := rm.newChangeBatch(action, recordSet)
	input.SetChangeBatch(changeBatch)
