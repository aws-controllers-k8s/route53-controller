
	// Setting the starting point to these values makes lookup faster
	if r.ko.Spec.Name != nil {
		input.SetStartRecordName(*r.ko.Spec.Name)
	}
	if r.ko.Spec.RecordType != nil {
		input.SetStartRecordType(*r.ko.Spec.RecordType)
	}
	if r.ko.Spec.SetIdentifier != nil {
		input.SetStartRecordIdentifier(*r.ko.Spec.SetIdentifier)
	}
