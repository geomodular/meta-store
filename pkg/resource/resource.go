package resource

import (
	"fmt"
	"strings"
)

type Resource struct {
	collectionName string
	resourceName   string
	subResource    *Resource
}

func New(collectionName, resourceName string) *Resource {
	return &Resource{
		collectionName,
		resourceName,
		nil,
	}
}

// NewFromPairs constructs Resource based on names, pairs of <collectionName, resourceName>.
// If there is an odd number of names, it sets the last name as resourceName.
func NewFromPairs(names ...string) *Resource {

	if len(names) == 0 {
		return nil
	}

	var rs []*Resource

	i := 0
	for _, name := range names {
		if len(rs) == i {
			rs = append(rs, &Resource{})
		}
		if rs[i].collectionName == "" && rs[i].resourceName == "" {
			rs[i].resourceName = name
		} else {
			rs[i].collectionName = rs[i].resourceName
			rs[i].resourceName = name
			i++
		}
	}

	last := rs[0]
	for _, r := range rs[1:] {
		last.subResource = r
		last = r
	}

	return rs[0]
}

// ParseResource takes a resource in a form of idiomatic path, parses it and returns a Resource.
func ParseResource(resourceName string) *Resource {
	s := strings.Split(resourceName, "/")
	return NewFromPairs(s...)
}

// CollectionName returns a collection name.
func (r Resource) CollectionName() string {
	return r.collectionName
}

// ResourceName returns a resource name.
func (r Resource) ResourceName() string {
	return r.resourceName
}

// deepCopy returns a deep copy of a resource.
func (r Resource) deepCopy() *Resource {
	var subResource *Resource

	if r.subResource != nil {
		subResource = r.subResource.deepCopy()
	}

	return &Resource{
		r.collectionName,
		r.resourceName,
		subResource,
	}
}

// Join merges two resources together. Immutable.
func (r Resource) Join(rs *Resource) *Resource {
	rt := r.deepCopy()

	var last *Resource
	for c := rt; c != nil; c = c.subResource {
		last = c
	}

	last.subResource = rs.deepCopy()

	return rt
}

// Add adds new sub-resource into existing resource. Immutable.
func (r Resource) Add(collectionName, resourceName string) *Resource {
	rt := r.deepCopy()

	var last *Resource
	for c := rt; c != nil; c = c.subResource {
		last = c
	}

	last.subResource = &Resource{
		collectionName,
		resourceName,
		nil,
	}

	return rt
}

// AddPairs adds new pairs <collectionName, resourceName> to resource the same way as NewFromPairs().
func (r Resource) AddPairs(names ...string) *Resource {
	rt := r.deepCopy()
	resource := NewFromPairs(names...)
	return rt.Join(resource)
}

// Unpack returns Resource as arrays of decoupled resources.
func (r Resource) Unpack() []*Resource {
	rt := []*Resource{
		{r.collectionName, r.resourceName, nil},
	}

	if r.subResource != nil {
		rt = append(rt, r.subResource.Unpack()...)
	}

	return rt
}

// FindByCollection returns a first match of a Resource with the same collectionName and all sub resources.
func (r Resource) FindByCollection(collectionName string) (*Resource, bool) {

	if r.collectionName == collectionName {
		return r.deepCopy(), true
	}

	if r.subResource != nil {
		return r.subResource.FindByCollection(collectionName)
	}

	return nil, false
}

// TrimRight removes an occurrence of resource `rs` for current resource `r` from the right side.
func (r Resource) TrimRight(rs *Resource) *Resource {
	rt := strings.TrimSuffix(r.String(), rs.String())
	rt = strings.TrimRight(rt, "/")
	return ParseResource(rt)
}

// String converts a resource into idiomatic id path.
func (r Resource) String() string {

	var subResource string
	if r.subResource != nil {
		subResource = r.subResource.String()
	}

	if r.collectionName == "" {
		if subResource != "" {
			return fmt.Sprintf("%s/%s", r.resourceName, subResource)
		}
		return r.resourceName
	}
	if subResource != "" {
		return fmt.Sprintf("%s/%s/%s", r.collectionName, r.resourceName, subResource)
	}
	return fmt.Sprintf("%s/%s", r.collectionName, r.resourceName)
}
