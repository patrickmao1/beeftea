package consensus

import (
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *Service) putMsg(msg *types.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outMsgs[string(msg.Hash())] = msg
}

// Takes s.outMsgs, wraps msgs into types.Envelope, sign it, and send to all peers
func (s *Service) processOutboundMsgs() {
	s.mu.Lock()
	defer s.mu.Unlock()
	//for id, msg := range s.outMsgs {
	//
	//}
}

func (s *Service) broadcast(msgs *types.Messages) error {
	//bs, err := proto.Marshal(msgs)
	//if err != nil {
	//	return err
	//}
	//sig := sign(s.MyKey(), bs)
	//for i, peer := range s.Peers {
	//
	//}
	return nil
}

func (s *Service) dialPeers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, peer := range s.Peers {
		dialOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
		cc, err := grpc.NewClient(peer.URL, dialOpt)
		if err != nil {
			log.Fatalf("failed to dial peer %d: %s", i, err.Error())
		}
		s.clients = append(s.clients, types.NewConsensusRPCClient(cc))
	}
}
