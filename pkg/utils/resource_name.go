package utils

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	MetaStoreResourceName = "metaStore"
	ServiceCollectionName = "services"
	DatasetCollectionName = "datasets"
)

func NewServiceResource(name string) *Resource {
	return NewResource(ServiceCollectionName, name)
}

func UnpackServiceResource(r *Resource) (string, error) {
	if r.collectionName == ServiceCollectionName {
		return r.resourceName, nil
	}
	return "", fmt.Errorf("not a service resource: %s", r)
}

func NewDatasetResource(collectionID uuid.UUID) *Resource {
	return NewResource(DatasetCollectionName, collectionID.String())
}

func NewParentResource() *Resource {
	return NewServiceResource(MetaStoreResourceName)
}

func UnpackDatasetResource(r *Resource) (uuid.UUID, error) {
	if r.collectionName == DatasetCollectionName {
		collectionID, err := uuid.Parse(r.resourceName)
		if err != nil {
			return uuid.UUID{}, errors.Wrapf(err, "failed parsing dataset uuid: %s", r.resourceName)
		}
		return collectionID, nil
	}
	return uuid.UUID{}, fmt.Errorf("not a project resource: %s", r)
}

func DatasetIDFromResourceName(resourceName string) (uuid.UUID, error) {
	r := ParseResource(resourceName)
	artifactResource, ok := r.FindByCollection(DatasetCollectionName)

	if ok {
		collectionID, err := UnpackDatasetResource(artifactResource)
		if err != nil {
			return uuid.UUID{}, errors.Wrap(err, "failed unpacking artifact resource")
		}
		return collectionID, nil
	}

	return uuid.UUID{}, errors.Errorf("no artifact in resource name: %s", resourceName)
}
