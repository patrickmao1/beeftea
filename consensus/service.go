package consensus

import (
	"encoding/binary"
	"github.com/patrickmao1/beeftea/crypto"
	"github.com/patrickmao1/beeftea/network"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
	"slices"
	"sync"
	"time"
)

type roundState struct {
	prevProposerProof []byte
	seed              []byte
	minProposal       *types.Proposal
	prepares          []*types.Prepare
	commits           []*types.Commit
}

type Service struct {
	*types.Config
	*roundState
	*network.Network

	mu sync.RWMutex

	// Operations requested by users
	reqs map[string]*types.PutReq // id -> PutReq

	// The key-value store
	db map[string]string
}

func NewService(config *types.Config) *Service {
	s := &Service{
		Config: config,
		reqs:   make(map[string]*types.PutReq),
	}
	s.Network = network.NewNetwork(
		config.MyIndex(),
		config.MyKey(),
		config.Peers,
		s.handleMessage,
	)
	return s
}

// Start starts the main consensus loop and the RPC servers
func (s *Service) Start() {
	log.Infoln("starting network")
	go s.Network.Start()

	log.Infoln("starting external RPC")
	go s.startRPC()

	log.Infoln("starting consensus main loop")

	timer := time.NewTicker(time.Until(s.roundEndTime()))
	for {
		<-timer.C
		proposalPhaseEnd := time.After(s.ProposalDuration)

		s.initRound()

		s.propose()

		// Blocks until the proposal phase has ended
		<-proposalPhaseEnd

		// Refresh round timer
		timer.Reset(time.Until(s.roundEndTime()))
	}
}

func (s *Service) initRound() {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := &roundState{}
	if s.roundState == nil {
		initSeed := blake2b.Sum256([]byte("beeftea"))
		state.prevProposerProof = initSeed[:]
	} else {
		// Use the latest known proposer proof as PP_{r-1}
		if s.minProposal == nil {
			state.prevProposerProof = s.prevProposerProof
		} else {
			state.prevProposerProof = state.minProposal.ProposerProof
		}
	}
	s.roundState = state
	s.seed = computeRoundSeed(s.round(), s.prevProposerProof)
	log.Infof("new round %d", s.round())
}

func computeRoundSeed(round uint32, prevPP []byte) (seed []byte) {
	// compute s_r, the seed for round r: s_r = r | PP_{r-1}
	// where PP_{r-1} is the proposer proof of the latest known proposer proof of the last round.
	// Mixing r into the calculation ensures that s_r changes for every r.
	// We use HASH("beeftea") as a PP_0.
	_seed := blake2b.Sum256(prevPP)
	seed = _seed[:]
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, round)
	return slices.Concat(b, seed[:])
}

func (s *Service) propose() {
	// For the ith peer
	// proposer proof: L_{i,r} = SIGN_i(s_r)
	// proposal score: S_{i,r} = HASH(L_{i,r})
	proposalScore, proposerProof := crypto.VRF(s.MyKey(), s.seed)
	if proposalScore >= s.ProposalThreshold {
		return
	}

	if len(s.reqs) == 0 {
		return
	}

	log.Infof("Proposing %d reqs", len(s.reqs))

	s.mu.Lock()
	reqs := make([]*types.PutReq, len(s.reqs))
	for _, req := range s.reqs {
		reqs = append(reqs, req)
	}
	proposal := &types.Proposal{
		Reqs:          reqs,
		ProposerProof: proposerProof,
		ProposerIndex: s.MyIndex(),
	}
	s.mu.Unlock()

	msg := &types.Message{Type: &types.Message_Proposal{
		Proposal: proposal,
	}}
	s.Broadcast(msg)
}

func (s *Service) prepare() {

}

func (s *Service) commit() {

}

func (s *Service) commitLocal() {

}

func (s *Service) round() uint32 {
	return uint32(time.Since(s.InitTime) / s.RoundDuration)
}

func (s *Service) roundEndTime() time.Time {
	elapsed := time.Duration(s.round()+1) * s.RoundDuration
	return s.InitTime.Add(elapsed)
}
