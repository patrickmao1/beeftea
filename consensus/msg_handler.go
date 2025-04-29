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
	// collect all proposals, then once the timer ends, call prepare with the minimum 
	// actually keep all proposals for later, edit roundstate so that it stores all proposals

	s.mu.Lock()
    defer s.mu.Unlock()

    // Make sure we are in the right round
    if proposal.Height != s.roundState.height || proposal.Round != s.roundState.round {
        log.Warnf("Received proposal for wrong round: got (h=%d, r=%d), expected (h=%d, r=%d)",
            proposal.Height, proposal.Round, s.roundState.height, s.roundState.round)
        return true, nil // Defer it: maybe we haven't advanced to this round yet
    }

	s.roundState.proposals = append(s.roundState.proposals, proposal)

    if s.roundState.minProposal == nil {
        s.roundState.minProposal = proposal
        log.Infof("Set initial minProposal to proposal with hash %x", HashProposal(proposal))
    } else {
        currentHash := HashProposal(s.roundState.minProposal)
        newHash := HashProposal(proposal)
        if bytes.Compare(newHash, currentHash) < 0 {
            s.roundState.minProposal = proposal
            log.Infof("Updated minProposal to proposal with smaller hash %x", newHash)
        } else {
            log.Infof("Ignored proposal with larger hash %x", newHash)
        }
    }

    return false, nil
}

func HashProposal(p *types.Proposal) []byte {
    data, err := proto.Marshal(p)
    if err != nil {
        log.Panicf("Failed to marshal proposal: %v", err)
    }
    hash := blake2b.Sum256(data)
    return hash[:]
}

func (s *Service) handlePrepare(prep *types.Prepare) (shouldDefer bool, err error) {
	// msg that i get here should be the same as minproposal
	// store all prepares in ether an array or a map, then check if we have reached quorum on any of the prepares
	// once we reach quorum, call commit
	return false, nil
}

func (s *Service) handleCommit(comm *types.Commit) (shouldDefer bool, err error) {
	return false, nil
}
