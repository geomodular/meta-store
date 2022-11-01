package artifact

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/geomodular/meta-store/pkg/log"
	"github.com/geomodular/meta-store/pkg/resource"
)

func Update[T any](ctx context.Context, db driver.Database, collectionName, name string, a *T) (*T, error) {

	key, err := resource.UUIDFromResourceName(name, collectionName)
	if err != nil {
		return nil, log.Report(err, "failed parsing artifact id")
	}

	col, err := db.Collection(ctx, collectionName)
	if err != nil {
		return nil, log.Report(err, "failed searching for collection")
	}

	var newArtifact T
	ctx = driver.WithReturnNew(ctx, &newArtifact)
	meta, err := col.UpdateDocument(ctx, key.String(), a)
	if err != nil {
		return nil, log.Report(err, "failed updating artifact")
	}

	log.ArangoMeta(meta, "artifact updated")

	return &newArtifact, nil
}
