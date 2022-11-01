package artifact

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/geomodular/meta-store/pkg/log"
	"github.com/geomodular/meta-store/pkg/resource"
	"github.com/google/uuid"
)

func Create[T any](ctx context.Context, db driver.Database, collectionName string, a Artifact) (*T, error) {

	key := uuid.New()

	serviceName := resource.NewMetaStoreResource()
	resourceName := serviceName.Join(resource.New(collectionName, key.String()))

	a.SetKey(key.String())
	a.SetName(resourceName.String())
	a.SetParent(serviceName.String())

	// Arango specific.

	col, err := db.Collection(ctx, collectionName)
	if err != nil {
		return nil, log.Report(err, "failed searching for collection")
	}

	var newArtifact T
	ctx = driver.WithReturnNew(ctx, &newArtifact)
	meta, err := col.CreateDocument(ctx, a)
	if err != nil {
		return nil, log.Report(err, "failed creating artifact")
	}

	log.ArangoMeta(meta, "new artifact created")

	return &newArtifact, nil
}
