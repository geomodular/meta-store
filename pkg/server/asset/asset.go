package asset

import (
	"github.com/arangodb/go-driver"
	"google.golang.org/grpc"
)

type grpcInitializer func(*grpc.Server, driver.Database)

var grpcInitializers []grpcInitializer

func RegisterGRPCInitializer(f grpcInitializer) {
	grpcInitializers = append(grpcInitializers, f)
}

func GetGRPCInitializers() []grpcInitializer {
	return grpcInitializers
}

var collections = map[string]struct{}{}

func RegisterCollection(c string) {
	collections[c] = struct{}{}
}

func GetCollections() []string {
	cs := make([]string, 0, len(collections))
	for k := range collections {
		cs = append(cs, k)
	}
	return cs
}
