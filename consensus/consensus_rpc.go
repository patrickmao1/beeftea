package consensus

import (
	"context"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

func (s *Service) startConsensusRPC() {
	svr := grpc.NewServer()
	lis, err := net.Listen("tcp", "0.0.0.0:9090")
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("grpc listening on 0.0.0.0:9090")
	types.RegisterConsensusRPCServer(svr, s)
	err = svr.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

// Send handles incoming call to the Send gRPC
func (s *Service) Send(ctx context.Context, envelope *types.Envelope) (*types.Empty, error) {
	//TODO implement me
	panic("implement me")
}
