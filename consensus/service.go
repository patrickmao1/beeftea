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
// list of the proposals, in commitlocal check individually if the digests match, when they do then putReq
// map for prepares and commits, it should have the digest as a key and the value is an array of the "from" field so
// that some node can't continuously send prepare messages and get quorum by itself
type roundState struct {
	prevProposerProof []byte
	seed              []byte
	minProposal       *types.Proposal
	proposals         []*types.Proposal
	prepares          map[string]map[uint32]bool
	commits           map[string]map[uint32]bool
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

		s.prepare()

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

// prepare implements phase 2: select the minimal valid proposal
// and broadcast a Prepare once it has seen 2f+1 matching proposals.
// prepare broadcasts a Prepare for the given proposal digest.
// prepare implements phase 2: take the chosen minProposal, compute its digest, and broadcast a Prepare message.
// It assumes s.minProposal is non-nil and has been set by handlePrepare.
func (s *Service) prepare() error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.minProposal == nil {
        return errors.New("no minProposal to prepare")
    }

    // compute full hash of minProposal
    raw, err := proto.Marshal(s.minProposal)
    if err != nil {
        return err
    }
    sum := blake2b.Sum256(raw)
    digest := sum[:] // full 32-byte digest

    // broadcast Prepare message carrying full digest
    pr := &types.Prepare{ProposalDigest: digest}
    msg := &types.Message{Type: &types.Message_Prepare{Prepare: pr}}
    s.Broadcast(msg)

    // track our local prepare for later commit
    s.prepares = append(s.prepares, pr)
    key := binary.BigEndian.Uint64(digest[:8])
    log.Infof("round %d: sent Prepare for digest %x", s.round(), key)
    return nil
}


// commit broadcasts a Commit for the given proposal digest.
func (s *Service) commit(proposalDigest []byte) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if len(proposalDigest) == 0 {
        return errors.New("empty proposal digest")
    }

    cm := &types.Commit{ProposalDigest: proposalDigest}
    msg := &types.Message{Type: &types.Message_Commit{Commit: cm}}
    s.Broadcast(msg)

    // track local commit
    s.commits = append(s.commits, cm)
    log.Infof("round %d: sent Commit for digest %x", s.round(), binary.BigEndian.Uint64(proposalDigest[:8]))
    return nil
}

func (s *Service) commitLocal(digest []byte) {
    for _, proposal := range s.roundState.proposals {
        if bytes.Equal(HashProposal(proposal), digest) {
            for _, req := range proposal.Reqs {
                s.kvStore[req.Key] = req.Value
            }
            break
        }
		

    }
}

func (s *Service) round() uint32 {
	return uint32(time.Since(s.InitTime) / s.RoundDuration)
}

func (s *Service) roundEndTime() time.Time {
	elapsed := time.Duration(s.round()+1) * s.RoundDuration
	return s.InitTime.Add(elapsed)
}
