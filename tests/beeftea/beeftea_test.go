package beeftea

import (
	"context"
	"errors"
	"fmt"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strconv"
	"sync"
	"testing"
	"time"
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
	num := "13"
	res, err := clients[0].Put(context.Background(), &types.PutReq{
		Id: num,
		Kv: &types.KeyValue{
			Key: "hello" + num,
			Val: "world" + num,
		},
	})
	require.NoError(t, err)
	log.Infof("res %+v", res)

	time.Sleep(4 * time.Second)

	key := "hello" + num
	log.Infof("querying value for key \"%s\"", key)
	val, err := getValue(key)
	require.NoError(t, err)
	log.Infof("val %s", val)
	require.EqualValues(t, "world"+num, val)
}

func TestGet(t *testing.T) {
	key := "hello11"
	log.Infof("querying value for key \"%s\"", key)
	val, err := getValue(key)
	require.NoError(t, err)
	log.Infof("val %s", val)
}

func TestPutMany(t *testing.T) {
	for i := 0; i < 10; i++ {
		res, err := clients[0].Put(context.Background(), &types.PutReq{
			Id: strconv.Itoa(i),
			Kv: &types.KeyValue{
				Key: fmt.Sprintf("hello%d", i),
				Val: fmt.Sprintf("world%d", i),
			},
		})
		require.NoError(t, err)
		log.Infof("res %+v", res)
	}
	time.Sleep(4 * time.Second)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("hello%d", i)
		val := fmt.Sprintf("world%d", i)
		log.Infof("querying value for key \"%s\"", key)
		res, err := getValue(key)
		require.NoError(t, err)
		log.Infof("val %s", val)
		require.EqualValues(t, val, res)
	}
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

	const quorum = 3
	if maxCount < quorum {
		return "", errors.New("no value has reached quorum")
	}

	return maxVal, nil
}
