package consensus

import (
	"github.com/patrickmao1/beeftea/types"
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
