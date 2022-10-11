package service

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	pb "github.com/geomodular/meta-store/gen/ai/h2o/meta_store"
	"github.com/geomodular/meta-store/pkg/utils"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

const (
	grpcEndpoint   = "localhost:9090"
	arangoEndpoint = "http://localhost:8529"
	arangoDB       = "metaStore"
)

type datasetSuite struct {
	suite.Suite

	datasetClient pb.DatasetServiceClient
	db            driver.Database
	datasetKeys   []string
}

func (s *datasetSuite) SetupSuite() {

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	grpcConn, err := grpc.Dial(grpcEndpoint, opts...)
	s.Require().NoError(err)

	datasetClient := pb.NewDatasetServiceClient(grpcConn)
	s.datasetClient = datasetClient

	dbConn, err := http.NewConnection(http.ConnectionConfig{Endpoints: []string{arangoEndpoint}})
	s.Require().NoError(err)

	client, err := driver.NewClient(driver.ClientConfig{Connection: dbConn})
	s.Require().NoError(err)

	db, err := client.Database(nil, arangoDB)
	s.Require().NoError(err)

	s.db = db
}

func (s *datasetSuite) TearDownTest() {
	ctx := context.Background()

	col, err := s.db.Collection(ctx, DatasetCollection)
	s.Require().NoError(err)

	_, _, err = col.RemoveDocuments(ctx, s.datasetKeys)
	s.Require().NoError(err)

	s.datasetKeys = nil
}

func (s *datasetSuite) TestCreateDataset() {
	ctx := context.Background()
	dataset_, err := s.datasetClient.CreateDataset(ctx, &pb.CreateDatasetRequest{
		Mime: "application/csv",
		Dataset: &pb.Dataset{
			Filename:    "iris.csv",
			Description: "Iris dataset",
			DisplayName: "Iris dataset",
		},
	})
	s.Require().NoError(err)

	key, err := utils.DatasetIDFromResourceName(dataset_.GetName())
	s.Require().NoError(err)

	col, err := s.db.Collection(ctx, DatasetCollection)
	s.Require().NoError(err)

	var dataset pb.MetaDataset
	_, err = col.ReadDocument(ctx, key.String(), &dataset)
	s.Require().NoError(err)

	s.Equal(dataset.Filename, "iris.csv")
	s.Equal(dataset.Description, "Iris dataset")
	s.Equal(dataset.DisplayName, "Iris dataset")
	s.Equal(dataset.Parent, "services/metaStore")

	// Clean up documents.
	s.datasetKeys = []string{key.String()}
}

func (s *datasetSuite) TestListDatasets() {

	ctx := context.Background()

	var keys []string
	for i := 0; i < 5; i++ {
		dataset_, err := s.datasetClient.CreateDataset(ctx, &pb.CreateDatasetRequest{
			Mime: "application/csv",
			Dataset: &pb.Dataset{
				Filename:    "iris.csv",
				Description: "Iris dataset",
				DisplayName: "Iris dataset",
			},
		})
		s.Require().NoError(err)

		key, err := utils.DatasetIDFromResourceName(dataset_.GetName())
		s.Require().NoError(err)

		keys = append(keys, key.String())
	}

	res, err := s.datasetClient.ListDatasets(ctx, &pb.ListDatasetsRequest{
		PageSize:  2,
		PageToken: "",
	})
	s.Require().NoError(err)
	s.EqualValues(2, res.TotalSize)
	s.EqualValues(2, len(res.Datasets))

	res2, err := s.datasetClient.ListDatasets(ctx, &pb.ListDatasetsRequest{
		PageSize:  0,
		PageToken: res.NextPageToken,
	})
	s.Require().NoError(err)
	s.EqualValues(2, res2.TotalSize)
	s.EqualValues(2, len(res2.Datasets))

	res3, err := s.datasetClient.ListDatasets(ctx, &pb.ListDatasetsRequest{
		PageSize:  0,
		PageToken: res2.NextPageToken,
	})
	s.Require().NoError(err)
	s.EqualValues(1, res3.TotalSize)
	s.EqualValues(1, len(res3.Datasets))
	s.Equal(res3.NextPageToken, "")

	// Clean up.
	s.datasetKeys = keys
}

func (s *datasetSuite) TestGetDataset() {
	ctx := context.Background()
	dataset_, err := s.datasetClient.CreateDataset(ctx, &pb.CreateDatasetRequest{
		Mime: "application/csv",
		Dataset: &pb.Dataset{
			Filename:    "iris2.csv",
			Description: "Iris2 dataset",
			DisplayName: "Iris2 dataset",
		},
	})
	s.Require().NoError(err)

	dataset, err := s.datasetClient.GetDataset(ctx, &pb.GetDatasetRequest{
		Name: dataset_.Name,
	})
	s.Require().NoError(err)

	s.Equal(dataset.Filename, "iris2.csv")
	s.Equal(dataset.Description, "Iris2 dataset")
	s.Equal(dataset.DisplayName, "Iris2 dataset")
	s.Equal(dataset.Parent, "services/metaStore")

	// Clean up documents.
	key, err := utils.DatasetIDFromResourceName(dataset_.GetName())
	s.Require().NoError(err)
	s.datasetKeys = []string{key.String()}
}

func TestDatasetSuite(t *testing.T) {
	suite.Run(t, new(datasetSuite))
}
