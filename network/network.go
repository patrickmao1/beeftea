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

	deferred map[string]*types.Message
}

type HandleMsgFunc func(msg *types.Message) (shouldDefer bool)

func NewNetwork(myIndex uint32, key *ecdsa.PrivateKey, handleMsg HandleMsgFunc) *Network {
	n := &Network{
		idx:       myIndex,
		key:       key,
		handleMsg: handleMsg,
		deferred:  make(map[string]*types.Message),
	}
	return n
}

func (n *Network) Start() {
	go n.startRPC()
	go n.dialPeers()
	n.processDeferred()
}

// Broadcast sends the msg to all nodes in the network asynchronously.
// NOTE: this function returns immediately without waiting for the other nodes to respond
func (n *Network) Broadcast(msg *types.Message) {
	n.doBroadcast([]*types.Message{msg})
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
		newRetries := make(map[string]*types.Message)
		for id, msg := range n.deferred {
			retry := n.handleMsg(msg)
			if retry {
				newRetries[id] = msg
			}
		}
		n.deferred = newRetries
		n.mu.Unlock()
	}
}

func (n *Network) ingestInbound(msgs []*types.Message) {
	for _, msg := range msgs {
		go n.handleMsg(msg)
	}
}

func (n *Network) doBroadcast(msgs []*types.Message) {
	_msgs := &types.Messages{Msgs: msgs}
	envelope := &types.Envelope{
		Msgs:      _msgs,
		NodeIndex: n.idx,
		Sig:       n.sign(_msgs),
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
