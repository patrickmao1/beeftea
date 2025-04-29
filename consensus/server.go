package consensus

import (
	"context"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

// Handles the incoming request from clients that wants to interact with the system
func (s *Service) startRPC() {
	svr := grpc.NewServer()
	lis, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("External server listening on 0.0.0.0:8080")
	types.RegisterExternalRPCServer(svr, s)
	err = svr.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Service) Put(ctx context.Context, req *types.PutReq) (*types.PutRes, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reqs[req.Id] = req
	return &types.PutRes{Id: req.Id}, nil
}

func (s *Service) Get(ctx context.Context, req *types.GetReq) (*types.GetRes, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val := s.db[req.Key]
	return &types.GetRes{Kv: &types.KeyValue{
		Key: req.Key,
		Val: val,
	}}, nil
}
