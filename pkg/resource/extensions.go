package resource

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	MetaStoreResourceName = "metaStore"
	ServiceCollectionName = "services"
)

func NewServiceResource(name string) *Resource {
	return New(ServiceCollectionName, name)
}

func UnpackServiceResource(r *Resource) (string, error) {
	if r.collectionName == ServiceCollectionName {
		return r.resourceName, nil
	}
	return "", fmt.Errorf("not a service resource: %s", r)
}

func NewMetaStoreResource() *Resource {
	return NewServiceResource(MetaStoreResourceName)
}

// TODO: func NewUUIDResource(collectionName string, collectionID uuid.UUID) *Resource

func UnpackUUIDResource(r *Resource, collectionName string) (uuid.UUID, error) {
	if r.collectionName == collectionName {
		collectionID, err := uuid.Parse(r.resourceName)
		if err != nil {
			return uuid.UUID{}, errors.Wrapf(err, "failed parsing dataset uuid: %s", r.resourceName)
		}
		return collectionID, nil
	}
	return uuid.UUID{}, fmt.Errorf("not a project resource: %s", r)
}

func UUIDFromResourceName(resourceName string, collectionName string) (uuid.UUID, error) {
	r := ParseResource(resourceName)
	artifactResource, ok := r.FindByCollection(collectionName)

	if ok {
		collectionID, err := UnpackUUIDResource(artifactResource, collectionName)
		if err != nil {
			return uuid.UUID{}, errors.Wrap(err, "failed unpacking artifact resource")
		}
		return collectionID, nil
	}

	return uuid.UUID{}, errors.Errorf("no artifact in resource name: %s", resourceName)
}
