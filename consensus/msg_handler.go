package consensus

import (
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
)

func (s *Service) handleMessage(msg *types.Message) (shouldDefer bool) {
	var err error
	switch msg.Type.(type) {
	case *types.Message_Proposal:
		shouldDefer, err = s.handleProposal(msg.GetProposal())
	case *types.Message_Prepare:
		shouldDefer, err = s.handlePrepare(msg.GetPrepare())
	case *types.Message_Commit:
		shouldDefer, err = s.handleCommit(msg.GetCommit())
	default:
		log.Panicf("unsupported message type: %T", msg.Type)
	}
	if err != nil {
		log.Errorf("handleMessage err: %s", err.Error())
	}
	return shouldDefer
}

func (s *Service) handleProposal(proposal *types.Proposal) (shouldDefer bool, err error) {
	return false, nil
}

func (s *Service) handlePrepare(prep *types.Prepare) (shouldDefer bool, err error) {
	return false, nil
}

func (s *Service) handleCommit(comm *types.Commit) (shouldDefer bool, err error) {
	return false, nil
}
