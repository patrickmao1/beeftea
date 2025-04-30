package consensus

import (
	"bytes"

	"github.com/patrickmao1/beeftea/crypto"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/protobuf/proto"
)

func (s *Service) handleMessage(nodeIdx uint32, msg *types.Message) (shouldDefer bool) {
	var err error
	switch msg.Type.(type) {
	case *types.Message_Proposal:
		shouldDefer, err = s.handleProposal(msg.GetProposal())
	case *types.Message_Prepare:
		shouldDefer, err = s.handlePrepare(msg.GetPrepare(), nodeIdx)
	case *types.Message_Commit:
		shouldDefer, err = s.handleCommit(msg.GetCommit(), nodeIdx)
	default:
		log.Panicf("unsupported message type: %T", msg.Type)
	}
	if err != nil {
		log.Errorf("handleMessage err: %s", err.Error())
	}
	return shouldDefer
}

// this method is called when the message is a proposal
func (s *Service) handleProposal(proposal *types.Proposal) (shouldDefer bool, err error) {
	// collect all proposals, then once the timer ends, call prepare with the minimum
	// actually keep all proposals for later, edit roundstate so that it stores all proposals

	s.mu.Lock()
	defer s.mu.Unlock()

	pubkey := &s.Peers[proposal.ProposerIndex].Key.PublicKey
	pass := crypto.Verify(pubkey, s.seed, proposal.ProposerProof)
	if !pass {
		log.Warnf("proposal from node %d verify fail", proposal.ProposerIndex)
		// verification failed maybe because I'm not in the same round (due to a bit of desync)
		// as the proposer, retry processing this proposal later.
		return true, nil
	}

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

func (s *Service) handlePrepare(prep *types.Prepare, nodeIdx uint32) (shouldDefer bool, err error) {
	// msg that i get here should be the same as minproposal
	// store all prepares in ether an array or a map, then check if we have reached quorum on any of the prepares
	// once we reach quorum, call commit
	s.mu.Lock()
	defer s.mu.Unlock()

	currentRound := s.round()

	if s.roundState == nil || currentRound != s.round() {
		log.Warnf("Deferring Prepare: node is not in the correct round (%d)", currentRound)
		return true, nil
	}

	if !s.roundState.prepared {
		log.Warnf("Deferring Prepare: haven't prepared yet, so can't accept others' prepares")
		return true, nil
	}

	digest := string(prep.ProposalDigest)

	// initialize map if digest is seen for the first time
	if s.roundState.prepares == nil {
		s.roundState.prepares = make(map[string]map[uint32]bool)
	}
	if _, exists := s.roundState.prepares[digest]; !exists {
		s.roundState.prepares[digest] = make(map[uint32]bool)
	}

	// Check for double-vote
	if s.roundState.prepares[digest][nodeIdx] {
		log.Warnf("Duplicate Prepare received from node %d for digest %x", nodeIdx, prep.ProposalDigest)
		return false, nil
	}

	// Record the prepare vote
	s.roundState.prepares[digest][nodeIdx] = true
	log.Infof("Accepted Prepare from node %d for digest %x", nodeIdx, prep.ProposalDigest)

	if len(s.roundState.prepares[digest]) > s.f*2 && !s.roundState.committed {
		log.Infof("Prepare quorum reached for digest %x, broadcasting Commit", prep.ProposalDigest)
		go s.commit(prep.ProposalDigest) // Call asynchronously to avoid deadlock
		s.roundState.committed = true
	}
	return false, nil
}

// this method is called when the message is a commit
func (s *Service) handleCommit(comm *types.Commit, nodeIdx uint32) (shouldDefer bool, err error) {
	return false, nil
}
