package beeftea

import (
	"context"
	"errors"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
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

func TestGet(t *testing.T) {
	key := "hello"
	log.Infof("querying value for key \"%s\"", key)
	val, err := getValue(key)
	require.NoError(t, err)
	log.Infof("val %s", val)
}

func getValue(key string) (string, error) {
	vals := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := client.Get(context.Background(), &types.GetReq{Key: key})
			if err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			vals[res.Kv.Val]++
		}()
	}
	wg.Wait()

	maxVal := ""
	maxCount := 0
	for val, count := range vals {
		if count > maxCount {
			maxVal = val
			maxCount = count
		}
	}

	const quorum = 2 // f+1
	if maxCount < quorum {
		return "", errors.New("no value has reached quorum")
	}

	return maxVal, nil
}
