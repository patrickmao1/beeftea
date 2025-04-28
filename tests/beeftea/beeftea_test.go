package beeftea

import (
	"context"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

var urls = []string{
	"localhost:8081",
	"localhost:8082",
	"localhost:8083",
	"localhost:8084",
	"localhost:8085",
}

var clients []types.ExternalRPCClient

func init() {
	for i, url := range urls {
		dialOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
		cc, err := grpc.NewClient(url, dialOpt)
		if err != nil {
			log.Fatalf("failed to dial peer %d: %s", i, err.Error())
		}
		clients = append(clients, types.NewExternalRPCClient(cc))
	}
}

func TestPut(t *testing.T) {
	res, err := clients[0].Put(context.Background(), &types.PutReq{
		Id: "1",
		Kv: &types.KeyValue{
			Key: "hello",
			Val: "world",
		},
	})
	if err != nil {
		require.NoError(t, err)
	}
	log.Infof("res %+v", res)
}
