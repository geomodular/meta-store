package artifact

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/geomodular/meta-store/pkg/log"
)

func List[T any](ctx context.Context, db driver.Database, collectionName, token string, size int) (string, int, []*T, error) {

	var pageIn *Paginator
	pageToken := token
	if pageToken != "" {
		var err error
		pageIn, err = parsePage(pageToken)
		if err != nil {
			return "", 0, nil, log.Report(err, "failed parsing page token")
		}
	} else {
		pageSize := size
		if pageSize <= 0 || pageSize > defaultPageSize {
			pageSize = defaultPageSize
		}
		pageIn = newPaginator(0, pageSize)
	}

	queryString := fmt.Sprintf("FOR d IN %s LIMIT %d, %d RETURN d", collectionName, pageIn.Offset, pageIn.Size)
	cursor, err := db.Query(ctx, queryString, nil)
	if err != nil {
		return "", 0, nil, log.Report(err, "failed querying database")
	}
	defer cursor.Close()

	var artifacts []*T
	for {
		var artifact T

		_, err := cursor.ReadDocument(ctx, &artifact)

		if driver.IsNoMoreDocuments(err) {
			break
		}

		if err != nil {
			return "", 0, nil, log.Report(err, "failed reading artifact")
		}

		artifacts = append(artifacts, &artifact)
	}

	outToken := ""
	if len(artifacts) == pageIn.Size {
		pageOut := newPaginator(pageIn.Offset+len(artifacts), pageIn.Size)
		outToken = pageOut.MustEncode()
	}

	return outToken, len(artifacts), artifacts, nil
}
