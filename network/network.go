package network

import (
	"context"
	"crypto/ecdsa"
	"github.com/golang/protobuf/proto"
	"github.com/patrickmao1/beeftea/crypto"
	"github.com/patrickmao1/beeftea/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

type Network struct {
	idx       uint32
	key       *ecdsa.PrivateKey
	handleMsg HandleMsgFunc
	peers     []*types.Peer
	clients   []types.ConsensusRPCClient

	mu sync.Mutex

	deferred map[string]*types.Envelope
}

type HandleMsgFunc func(nodeIdx uint32, m *types.Message) (shouldDefer bool)

func NewNetwork(
	myIndex uint32,
	key *ecdsa.PrivateKey,
	peers []*types.Peer,
	handleMsg HandleMsgFunc,
) *Network {
	n := &Network{
		idx:       myIndex,
		key:       key,
		handleMsg: handleMsg,
		peers:     peers,
		deferred:  make(map[string]*types.Envelope),
	}
	return n
}

func (n *Network) Start() {
	log.Info("starting network, my peer index: ", n.idx)
	go n.startRPC()
	go n.dialPeers()
	n.processDeferred()
}

// Broadcast sends the msg to all nodes in the network asynchronously.
// NOTE: this function returns immediately without waiting for the other nodes to respond
func (n *Network) Broadcast(msg *types.Message) {
	n.doBroadcast(msg)
}

func (n *Network) dialPeers() {
	for i, peer := range n.peers {
		dialOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
		cc, err := grpc.NewClient(peer.URL, dialOpt)
		if err != nil {
			log.Errorf("failed to dial peer %d: %s", i, err.Error())
			continue
		}
		n.clients = append(n.clients, types.NewConsensusRPCClient(cc))
	}
}

func (n *Network) processDeferred() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		<-ticker.C
		n.mu.Lock()
		newRetries := make(map[string]*types.Envelope)
		for id, e := range n.deferred {
			shouldDefer := n.handleMsg(e.NodeIndex, e.Msg)
			if shouldDefer {
				newRetries[id] = e
			}
		}
		n.deferred = newRetries
		n.mu.Unlock()
	}
}

func (n *Network) ingest(e *types.Envelope) {
	go func() {
		shouldDefer := n.handleMsg(e.NodeIndex, e.Msg)
		if shouldDefer {
			n.mu.Lock()
			n.deferred[string(e.Hash())] = e
			n.mu.Unlock()
		}
	}()
}

func (n *Network) doBroadcast(msg *types.Message) {
	envelope := &types.Envelope{
		Msg:       msg,
		NodeIndex: n.idx,
		Sig:       n.sign(msg),
	}
	for i, client := range n.clients {
		go func() {
			_, err := client.Send(context.Background(), envelope)
			if err != nil {
				log.Errorf("failed to send to peer %d: %s", i, err.Error())
			}
		}()
	}

	return
}

func (n *Network) sign(msg proto.Message) []byte {
	bs, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return crypto.Sign(n.key, bs)
}
