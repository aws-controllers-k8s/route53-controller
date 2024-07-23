	ko := r.ko.DeepCopy()
	ko.Spec.Locations = nil
	desired := &resource{ko}
	_, err = rm.customUpdateCidrCollection(ctx, desired, r, nil)
	if err != nil {
		return nil, err
	}
	input.SetId(*r.ko.Status.Collection.ID)
