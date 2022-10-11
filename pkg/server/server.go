package server

import (
	"github.com/geomodular/meta-store/pkg/service"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	pb "github.com/geomodular/meta-store/gen/ai/h2o/meta_store"
	"google.golang.org/grpc"
)

const (
	grpcBind       = ":9090"
	arangoEndpoint = "http://localhost:8529"
	arangoDB       = "metaStore"
)

func Run() error {

	db, err := initArango(arangoEndpoint, arangoDB)
	if err != nil {
		return errors.Wrap(err, "failed initializing database")
	}
	log.Info().Msgf("connected to arango on %s", arangoEndpoint)

	grpcServer, err := initGRPC(db)
	if err != nil {
		return errors.Wrap(err, "failed initializing gRPC server")
	}

	lis, err := net.Listen("tcp", grpcBind)
	if err != nil {
		return errors.Wrapf(err, "failed listening on gRPC port %s", grpcBind)
	}

	log.Info().Msgf("starting gRPC server on %s", grpcBind)
	if err = grpcServer.Serve(lis); err != nil {
		log.Err(err).Msg("failed starting gRPC server")
	}

	return nil
}

func initArango(endpoint, dbName string) (driver.Database, error) {
	conn, err := http.NewConnection(http.ConnectionConfig{Endpoints: []string{endpoint}})
	if err != nil {
		return nil, errors.Wrap(err, "failed connecting to arangodb")
	}

	client, err := driver.NewClient(driver.ClientConfig{Connection: conn})
	if err != nil {
		return nil, errors.Wrap(err, "failed creating a client")
	}

	dbExists, err := client.DatabaseExists(nil, dbName)
	if err != nil {
		return nil, errors.Wrap(err, "failed checking for db existence")
	}

	var db driver.Database
	if !dbExists {
		db, err = client.CreateDatabase(nil, dbName, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed creating database")
		}
	} else {
		db, err = client.Database(nil, dbName)
		if err != nil {
			return nil, errors.Wrap(err, "failed opening database")
		}
	}

	// TODO: This should be checked and created based on the existing artifacts.
	// Also CreateCollectionOptions are important.
	colExists, err := db.CollectionExists(nil, service.DatasetCollection)
	if err != nil {
		return nil, errors.Wrap(err, "failed checking for collection existence")
	}

	if !colExists {
		_, err = db.CreateCollection(nil, service.DatasetCollection, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed creating collection")
		}
	}

	return db, nil
}

func initGRPC(db driver.Database) (*grpc.Server, error) {

	datasetServer := service.NewDatasetServer(db)

	grpcServer := grpc.NewServer()
	pb.RegisterDatasetServiceServer(grpcServer, datasetServer)
	return grpcServer, nil
}
