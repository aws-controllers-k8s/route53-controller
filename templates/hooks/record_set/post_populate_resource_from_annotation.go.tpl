	if f1, f1ok := fields["recordType"]; f1ok {
		r.ko.Spec.RecordType = &f1
	}

	if f2, f2ok := fields["name"]; f2ok {
		r.ko.Spec.Name = &f2
	}
