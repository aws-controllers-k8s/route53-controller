	f1, f1ok := identifier.AdditionalKeys["recordType"]
	if f1ok {
		r.ko.Spec.RecordType = &f1
	}

	if f2, f2ok := identifier.AdditionalKeys["name"]; f2ok {
		r.ko.Spec.Name = &f2
	}
