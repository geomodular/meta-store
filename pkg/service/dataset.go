package service

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/geomodular/meta-store/gen/ai/h2o/meta_store"
	"github.com/geomodular/meta-store/pkg/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	DatasetCollection = "datasets"
)

type datasetServer struct {
	db driver.Database
}

func NewDatasetServer(db driver.Database) *datasetServer {
	return &datasetServer{db: db}
}

func (d *datasetServer) CreateDataset(ctx context.Context, req *pb.CreateDatasetRequest) (*pb.Dataset, error) {

	dataset := pb.NewMetaDatasetFromProto(req.GetDataset())

	parentResource := utils.NewParentResource()

	key := uuid.New()
	dataset.Key = key.String()
	dataset.Name = parentResource.Join(utils.NewDatasetResource(key)).String()
	dataset.Parent = parentResource.String()
	dataset.CreateTime = time.Now()
	dataset.UpdateTime = time.Now()

	col, err := d.db.Collection(ctx, utils.DatasetCollectionName)
	if err != nil {
		return nil, report(err, "failed searching for collection")
	}

	var newDataset pb.MetaDataset
	ctx = driver.WithReturnNew(ctx, &newDataset)
	meta, err := col.CreateDocument(ctx, dataset)
	if err != nil {
		return nil, report(err, "failed creating document")
	}

	logMeta(meta, "new dataset created")

	return newDataset.ToProto(), nil
}

func (d *datasetServer) GetDataset(ctx context.Context, req *pb.GetDatasetRequest) (*pb.Dataset, error) {

	key, err := utils.DatasetIDFromResourceName(req.GetName())
	if err != nil {
		return nil, report(err, "failed parsing dataset id")
	}

	col, err := d.db.Collection(ctx, DatasetCollection)
	if err != nil {
		return nil, report(err, "failed searching for collection")
	}

	var dataset pb.MetaDataset
	meta, err := col.ReadDocument(ctx, key.String(), &dataset)
	if err != nil {
		return nil, report(err, "failed reading dataset in collection")
	}

	logMeta(meta, "dataset returned")

	return dataset.ToProto(), nil
}

func (d *datasetServer) ListDatasets(ctx context.Context, req *pb.ListDatasetsRequest) (*pb.ListDatasetsResponse, error) {

	var pageIn *Paginator
	pageToken := req.GetPageToken()
	if pageToken != "" {
		var err error
		pageIn, err = parsePage(pageToken)
		if err != nil {
			return nil, report(err, "failed parsing page token")
		}
	} else {
		pageSize := int(req.GetPageSize())
		if pageSize <= 0 || pageSize > defaultPageSize {
			pageSize = defaultPageSize
		}
		pageIn = newPaginator(0, pageSize)
	}

	log.Debug().Int("offset", pageIn.Offset).Int("size", pageIn.Size).Msgf("pagination info")

	// TODO: Should go elsewhere:
	queryString := fmt.Sprintf("FOR d IN datasets LIMIT %d, %d RETURN d", pageIn.Offset, pageIn.Size)
	cursor, err := d.db.Query(ctx, queryString, nil)
	if err != nil {
		return nil, report(err, "failed querying database")
	}
	defer cursor.Close()

	var datasets []*pb.Dataset
	for {
		var dataset pb.MetaDataset

		meta, err := cursor.ReadDocument(ctx, &dataset)

		if driver.IsNoMoreDocuments(err) {
			break
		}

		if err != nil {
			return nil, report(err, "failed reading document")
		}

		datasets = append(datasets, dataset.ToProto())

		logMeta(meta, "dataset being listed")
	}

	log.Debug().Int("documents found", len(datasets)).Msg("listed documents info")

	token := ""
	if len(datasets) == pageIn.Size {
		pageOut := newPaginator(pageIn.Offset+len(datasets), pageIn.Size)
		token = pageOut.MustEncode()
	}

	return &pb.ListDatasetsResponse{
		TotalSize:     int32(len(datasets)),
		NextPageToken: token,
		Datasets:      datasets,
	}, nil
}
