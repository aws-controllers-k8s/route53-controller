package cidr_collection

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/route53"
)

// getCallerReference will generate a CallerReference for a given health check
// using the current timestamp, so that it produces a unique value
func getCallerReference() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

// oldLocations returns a slice of CidrLocation pointer objects
// to delete.
func (rm *resourceManager) oldLocations(
	latest *resource,
) []*svcsdk.CidrCollectionChange {
	if latest == nil {
		return nil
	}

	locations := latest.ko.Spec.Locations
	var oldLocations []*svcsdk.CidrCollectionChange
	// Add removed locations to Changes using the DELETE_IF_EXISTS action
	if len(locations) > 0 {
		action := "DELETE_IF_EXISTS"
		for _, rr := range locations {
			change := &svcsdk.CidrCollectionChange{}
			change.SetLocationName(*rr.LocationName)
			change.SetCidrList(rr.CIDRList)
			change.SetAction(action)
			oldLocations = append(oldLocations, change)
		}
	}
	return oldLocations
}

// newLocations returns a slice of CidrLocation pointer objects
// with values set by the resource's corresponding spec field.
func (rm *resourceManager) newLocations(
	desired *resource,
) ([]*svcsdk.CidrCollectionChange, error) {
	if desired == nil {
		return nil, nil
	}

	locations := desired.ko.Spec.Locations

	var newLocations []*svcsdk.CidrCollectionChange
	// Add locations to Changes using the PUT action
	if len(locations) > 0 {
		action := "PUT"
		for _, rr := range locations {
			if rr.LocationName == nil {
				return nil, errors.New("InvalidInput: locationName is required in cidrLocation")
			}
			if rr.CIDRList == nil {
				return nil, errors.New("InvalidInput: cidrList is required in cidrLocation")
			}
			change := &svcsdk.CidrCollectionChange{}
			change.SetLocationName(*rr.LocationName)
			change.SetCidrList(rr.CIDRList)
			change.SetAction(action)
			newLocations = append(newLocations, change)
		}
	}

	return newLocations, nil
}

// customUpdateRecordSet is the custom implementation for
// RecordSet resource's update operation.
func (rm *resourceManager) customUpdateCidrCollection(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateCidrCollection")
	defer func() {
		exit(err)
	}()

	if delta != nil {
		// Do not proceed with update if an immutable field was updated
		if immutableFieldChanges := rm.getImmutableFieldChanges(delta); len(immutableFieldChanges) > 0 {
			msg := fmt.Sprintf("Immutable Spec fields have been modified: %s", strings.Join(immutableFieldChanges, ","))
			return nil, ackerr.NewTerminalError(fmt.Errorf(msg))
		}
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	input := &svcsdk.ChangeCidrCollectionInput{}
	input.SetId(*desired.ko.Status.Collection.ID)
	collectionVersion := *desired.ko.Status.Collection.Version
	input.SetCollectionVersion(collectionVersion)

	newLocations, err := rm.newLocations(desired)
	if err != nil {
		return nil, err
	}
	oldLocations := rm.oldLocations(latest)

	// First remove old Locations as there is no api call for updating cidr Collection locations
	if oldLocations != nil {
		input.SetChanges(oldLocations)
		_, err = rm.sdkapi.ChangeCidrCollectionWithContext(ctx, input)
		rm.metrics.RecordAPICall("UPDATE", "ChangeCidrCollection", err)
		if err != nil {
			return nil, err
		}
		collectionVersion = collectionVersion + 1
		input.SetCollectionVersion(collectionVersion)
		ko.Status.Collection.Version = &collectionVersion
	}

	// Add all new Locations to the cidr Collection
	if newLocations != nil {
		input.SetChanges(newLocations)
		_, err = rm.sdkapi.ChangeCidrCollectionWithContext(ctx, input)
		rm.metrics.RecordAPICall("UPDATE", "ChangeCidrCollection", err)
		if err != nil {
			return nil, err
		}
		collectionVersion = collectionVersion + 1
		input.SetCollectionVersion(collectionVersion)
		ko.Status.Collection.Version = &collectionVersion
	}

	return &resource{ko}, nil
}

// newListCidrLocationsRequestPayload returns SDK-specific struct for the HTTP request
// payload of the ListCidrBlocks API call for the resource
func (rm *resourceManager) newListCidrLocationsRequestPayload(
	r *resource,
) (*svcsdk.ListCidrLocationsInput, error) {
	res := &svcsdk.ListCidrLocationsInput{}

	return res, nil
}

// newListCidrBlocksRequestPayload returns SDK-specific struct for the HTTP request
// payload of the ListCidrBlocks API call for the resource
func (rm *resourceManager) newListCidrBlocksRequestPayload(
	r *resource,
) (*svcsdk.ListCidrBlocksInput, error) {
	res := &svcsdk.ListCidrBlocksInput{}

	return res, nil
}

// compareCustom is a custom comparison function for comparing Locations Slices
func compareCustom(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if len(a.ko.Spec.Locations) != len(b.ko.Spec.Locations) {
		delta.Add("Spec.Locations", a.ko.Spec.Locations, b.ko.Spec.Locations)
	} else if len(a.ko.Spec.Locations) > 0 {
		for index, elem := range a.ko.Spec.Locations {
			if elem.LocationName != b.ko.Spec.Locations[index].LocationName {
				delta.Add("Spec.Locations", a.ko.Spec.Locations, b.ko.Spec.Locations)
			}
			if !reflect.DeepEqual(elem.CIDRList, b.ko.Spec.Locations[index].CIDRList) {
				delta.Add("Spec.Locations", a.ko.Spec.Locations, b.ko.Spec.Locations)
			}
		}
	}
}
