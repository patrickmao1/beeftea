package consensus

import (
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
)

func (s *Service) handleMessage(msg *types.Message) {
	var err error
	switch msg.Type.(type) {
	case *types.Message_Proposal:
		err = s.handleProposal(msg.GetProposal())
	case *types.Message_Prepare:
		err = s.handlePrepare(msg.GetPrepare())
	case *types.Message_Commit:
		err = s.handleCommit(msg.GetCommit())
	}
	if err != nil {
		log.Errorf("handleMessage err: %s", err.Error())
	}
}

func (s *Service) handleProposal(proposal *types.Proposal) error {
	return nil
}

func (s *Service) handlePrepare(prep *types.Prepare) error {
	return nil
}

func (s *Service) handleCommit(comm *types.Commit) error {
	return nil
}
