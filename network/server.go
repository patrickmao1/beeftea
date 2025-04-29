package network

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/patrickmao1/beeftea/crypto"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

func (n *Network) startRPC() {
	svr := grpc.NewServer()
	lis, err := net.Listen("tcp", "0.0.0.0:9090")
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("grpc listening on 0.0.0.0:9090")
	types.RegisterConsensusRPCServer(svr, n)
	err = svr.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

// Send handles incoming call to the Send gRPC
func (n *Network) Send(_ context.Context, envelope *types.Envelope) (*types.Empty, error) {
	bs, err := proto.Marshal(envelope.Msgs)
	if err != nil {
		log.Errorf("failed to marshal msg: %s", err.Error())
		return nil, err
	}

	pubkey := n.peers[envelope.NodeIndex].Key.PublicKey
	pass := crypto.Verify(&pubkey, bs, envelope.Sig)
	if !pass {
		log.Errorf("failed to verify signature for msg from peer %d", envelope.NodeIndex)
	}

	n.ingestInbound(envelope.Msgs.Msgs)

	return &types.Empty{}, nil
}
