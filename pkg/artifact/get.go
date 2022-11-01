package artifact

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/geomodular/meta-store/pkg/log"
	"github.com/geomodular/meta-store/pkg/resource"
)

func Get[T any](ctx context.Context, db driver.Database, name string) (*T, error) {

	collectionName := "datasets" // TODO

	key, err := resource.UUIDFromResourceName(name, collectionName)
	if err != nil {
		return nil, log.Report(err, "failed parsing dataset id")
	}

	col, err := db.Collection(ctx, collectionName)
	if err != nil {
		return nil, log.Report(err, "failed searching for collection")
	}

	var dataset T
	meta, err := col.ReadDocument(ctx, key.String(), &dataset)
	if err != nil {
		return nil, log.Report(err, "failed reading dataset in collection")
	}

	log.ArangoMeta(meta, "dataset returned")

	return &dataset, nil
}
