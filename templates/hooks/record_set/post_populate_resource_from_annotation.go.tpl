	// If the pre hook injected an empty "id" sentinel (absent from annotation),
	// clear Status.ID so that sdkFind skips the ChangeInfo lookup.
	if r.ko.Status.ID != nil && *r.ko.Status.ID == "" {
		r.ko.Status.ID = nil
	}

	if f1, f1ok := fields["recordType"]; f1ok {
		r.ko.Spec.RecordType = &f1
	}

	if f2, f2ok := fields["name"]; f2ok {
		r.ko.Spec.Name = &f2
	}
