package asset

import (
	"github.com/arangodb/go-driver"
	"google.golang.org/grpc"
)

// TODO: Make SETs instead of arrays.

var grpcInitializers []func(*grpc.Server, driver.Database)

func RegisterGRPCInitializer(i func(*grpc.Server, driver.Database)) {
	grpcInitializers = append(grpcInitializers, i)
}

func GetGRPCInitializers() []func(*grpc.Server, driver.Database) {
	return grpcInitializers
}

var collections []string

func RegisterCollection(c string) {
	collections = append(collections, c)
}

func GetCollections() []string {
	return collections
}
