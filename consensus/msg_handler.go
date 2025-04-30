package consensus

import (
	"fmt"
	"github.com/patrickmao1/beeftea/crypto"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
)

func (s *Service) handleMessage(nodeIdx uint32, msg *types.Message) (shouldDefer bool) {
	var err error
	switch msg.Type.(type) {
	case *types.Message_Proposal:
		shouldDefer, err = s.handleProposal(msg.GetProposal(), nodeIdx)
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
func (s *Service) handleProposal(proposal *types.Proposal, nodeIdx uint32) (shouldDefer bool, err error) {
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
	s.roundState.proposals = append(s.roundState.proposals, proposal)

	newScore := proposal.Score()
	if newScore >= s.ProposalThreshold {
		return false, fmt.Errorf("received proposal score too big %d > %d", newScore, s.ProposalThreshold)
	}

	if s.roundState.minProposal == nil {
		s.roundState.minProposal = proposal
		log.Infof("Set initial minProposal to proposal with score %d", newScore)
	} else {
		curMinScore := s.minProposal.Score()
		if newScore < curMinScore {
			s.roundState.minProposal = proposal
			log.Infof("Updated minProposal to proposal with smaller score %d < %d", newScore, curMinScore)
		} else {
			log.Infof("Ignored proposal with larger score %d > %d", newScore, curMinScore)
		}
	}

	return false, nil
}

func (s *Service) handlePrepare(prep *types.Prepare, nodeIdx uint32) (shouldDefer bool, err error) {
	// msg that i get here should be the same as minproposal
	// store all prepares in ether an array or a map, then check if we have reached quorum on any of the prepares
	// once we reach quorum, call commit
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.roundState == nil || s.roundState.round != s.round() {
		log.Warnf("Deferring Prepare: not in the correct round (expected %d, got %d)", s.roundState.round, s.round())
		return true, nil
	}

	if !s.roundState.prepared {
		log.Warnf("Deferring Prepare: haven't prepared yet, so can't accept others' prepares")
		return true, nil
	}

	digest := string(prep.ProposalDigest[:8])

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
	log.Infof("Accepted Prepare from node %d for digest %x, current count %d",
		nodeIdx, prep.ProposalDigest, len(s.roundState.prepares[digest]))

	if len(s.roundState.prepares[digest]) >= 3 && !s.roundState.committed {
		log.Infof("Prepare quorum reached for digest %x, broadcasting Commit", prep.ProposalDigest)
		go func() {
			err := s.commit(prep.ProposalDigest) // Call asynchronously to avoid deadlock
			if err != nil {
				log.Errorf("commit failed: digest %x, err %s", prep.ProposalDigest, err.Error())
			}
		}()
		s.roundState.committed = true
	}
	return false, nil
}

// this method is called when the message is a commit
func (s *Service) handleCommit(comm *types.Commit, nodeIdx uint32) (shouldDefer bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.roundState == nil || s.roundState.round != s.round() {
		log.Warnf("Deferring Commit: not in the correct round (expected %d, got %d)", s.roundState.round, s.round())
		return true, nil
	}

	if !s.roundState.committed {
		log.Warnf("Deferring Commit: haven't committed yet, can't accept others' commits")
		return true, nil
	}

	digest := string(comm.ProposalDigest[:8])

	// Initialize commit map if needed
	if s.roundState.commits == nil {
		s.roundState.commits = make(map[string]map[uint32]bool)
	}
	if _, exists := s.roundState.commits[digest]; !exists {
		s.roundState.commits[digest] = make(map[uint32]bool)
	}

	// Prevent duplicate votes from the same node
	if s.roundState.commits[digest][nodeIdx] {
		log.Warnf("Duplicate Commit received from node %d for digest %x", nodeIdx, comm.ProposalDigest)
		return false, nil
	}

	// Record the vote
	s.roundState.commits[digest][nodeIdx] = true
	log.Infof("Accepted Commit from node %d for digest %x", nodeIdx, comm.ProposalDigest)

	// Quorum reached: finalize the decision
	if len(s.roundState.commits[digest]) >= 3 {
		log.Infof("Commit quorum reached for digest %x. Finalizing commit.", comm.ProposalDigest)
		go s.commitLocal(comm.ProposalDigest) // Call asynchronously to apply state changes
	}
	return false, nil
}
