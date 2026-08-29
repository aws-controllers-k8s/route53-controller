
	// id maps to Status.ID (ChangeInfo ID). It is intentionally optional for
	// adoption of pre-existing records, which have no associated ChangeInfo.
	// Inject an empty string so that the generated required-field check passes;
	// the post hook will clear Status.ID when the injected empty value is found.
	if _, ok := fields["id"]; !ok {
		fields["id"] = ""
	}
