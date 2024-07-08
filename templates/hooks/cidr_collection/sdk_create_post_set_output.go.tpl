	ko.Status.CallerReference = &callerReference
	updatedResource, err := rm.customUpdateCidrCollection(ctx, &resource{ko}, nil, nil)
	if err != nil {
		return nil, err
	}
	ko = updatedResource.ko.DeepCopy()
