package artifact

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/geomodular/meta-store/pkg/log"
	"github.com/geomodular/meta-store/pkg/resource"
)

func Delete(ctx context.Context, db driver.Database, collectionName, name string) error {

	key, err := resource.UUIDFromResourceName(name, collectionName)
	if err != nil {
		return log.Report(err, "failed parsing artifact id")
	}

	col, err := db.Collection(ctx, collectionName)
	if err != nil {
		return log.Report(err, "failed searching for collection")
	}

	meta, err := col.RemoveDocument(ctx, key.String())
	if err != nil {
		return log.Report(err, "failed removing artifact in collection")
	}

	log.ArangoMeta(meta, "artifact removed")

	return nil
}
