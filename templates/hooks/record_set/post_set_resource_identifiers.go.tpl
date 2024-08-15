	f1, f1ok := identifier.AdditionalKeys["recordType"]
	if f1ok {
		r.ko.Spec.RecordType = &f1
	}