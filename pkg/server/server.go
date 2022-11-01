package server

import (
	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"

	_ "github.com/geomodular/meta-store/gen/ai/h2o/meta_store"
	"github.com/geomodular/meta-store/pkg/server/asset"
)

const (
	grpcBind       = ":9090"
	arangoEndpoint = "http://root:openSesame@localhost:8529"
	arangoDB       = "metaStore"
)

func Run() error {

	db, err := initArango(arangoEndpoint, arangoDB)
	if err != nil {
		return errors.Wrap(err, "failed initializing database")
	}
	log.Info().Msgf("connected to arango on %s", arangoEndpoint)

	grpcServer := initGRPC(db)

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

	// Cycle through the registered collections.
	collections := asset.GetCollections()
	for _, collection := range collections {
		colExists, err := db.CollectionExists(nil, collection)
		if err != nil {
			return nil, errors.Wrap(err, "failed checking for collection existence")
		}

		if !colExists {
			_, err = db.CreateCollection(nil, collection, nil)
			if err != nil {
				return nil, errors.Wrap(err, "failed creating collection")
			}
		}
	}

	return db, nil
}

func initGRPC(db driver.Database) *grpc.Server {

	grpcServer := grpc.NewServer()

	// Cycle through the auto-generated servers.
	inits := asset.GetGRPCInitializers()
	for _, init := range inits {
		init(grpcServer, db)
	}

	return grpcServer
}
