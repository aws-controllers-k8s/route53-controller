
	// ListCidrCollections does not result in an exact match of relevant cidr locations as
	// there is no filter option, so we need to filter here
    // Furthermore we need different api calls to get the cidrLocations and cidrBlocks associated with the cidr Collection
    var resourceId string
    if ko.Status.Collection != nil {
        if ko.Status.Collection.ID != nil {
          resourceId = *ko.Status.Collection.ID
        }
    }

	var cidrCollections []*svcsdk.CollectionSummary
	for _, elem := range resp.CidrCollections {
		elemId := *elem.Id
		if elemId == resourceId {
			cidrCollections = append(cidrCollections, elem)
			if ko.Status.Collection == nil {
				ko.Status.Collection = &svcapitypes.CIDRCollection_SDK{}
			}
			ko.Status.Collection.ARN = elem.Arn
			ko.Status.Collection.ID = elem.Id
			ko.Status.Collection.Name = elem.Name
			ko.Status.Collection.Version = elem.Version
		}
	}

	if len(cidrCollections) == 0 {
		return nil, ackerr.NotFound
	}

	inputListCidrLocations, err := rm.newListCidrLocationsRequestPayload(r)
	if err != nil {
		return nil, err
	}
	inputListCidrLocations.SetCollectionId(*ko.Status.Collection.ID)
	respListCidrLocations, err := rm.sdkapi.ListCidrLocationsWithContext(ctx, inputListCidrLocations)
	rm.metrics.RecordAPICall("READ_MANY", "ListCidrLocations", err)
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "UNKNOWN" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	var locations []*svcapitypes.CIDRCollectionChange
	for _, elemCidrLocation := range respListCidrLocations.CidrLocations {
		location := svcapitypes.CIDRCollectionChange{}
		location.LocationName = elemCidrLocation.LocationName

		inputListCidrBlocks, err := rm.newListCidrBlocksRequestPayload(r)
		if err != nil {
			return nil, err
		}
		inputListCidrBlocks.SetCollectionId(*ko.Status.Collection.ID)
		inputListCidrBlocks.SetLocationName(*elemCidrLocation.LocationName)
		respListCidrBlocks, err := rm.sdkapi.ListCidrBlocksWithContext(ctx, inputListCidrBlocks)
		rm.metrics.RecordAPICall("READ_MANY", "ListCidrBlocks", err)
		if err != nil {
			if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "UNKNOWN" {
				return nil, ackerr.NotFound
			}
			return nil, err
		}

		var cidrList []*string
		for _, elemcidrBlock := range respListCidrBlocks.CidrBlocks {
			cidrList = append(cidrList, elemcidrBlock.CidrBlock)
		}
		location.CIDRList = cidrList
		locations = append(locations, &location)
	}
	if ko.Spec.Locations == nil {
		ko.Spec.Locations = []*svcapitypes.CIDRCollectionChange{}
	}
	ko.Spec.Locations = locations

	resp.CidrCollections = cidrCollections
