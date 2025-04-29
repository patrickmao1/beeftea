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

	mu           sync.Mutex
	inboundCond  *sync.Cond
	outboundCond *sync.Cond

	inbound  map[string]*types.Message
	deferred map[string]*types.Message
	outbound map[string]*types.Message
}

type HandleMsgFunc func(msg *types.Message) (shouldDefer bool)

func NewNetwork(myIndex uint32, key *ecdsa.PrivateKey, handleMsg HandleMsgFunc) *Network {
	n := &Network{
		idx:       myIndex,
		key:       key,
		handleMsg: handleMsg,
		inbound:   make(map[string]*types.Message),
		deferred:  make(map[string]*types.Message),
		outbound:  make(map[string]*types.Message),
	}
	n.inboundCond = sync.NewCond(&n.mu)
	n.outboundCond = sync.NewCond(&n.mu)
	return n
}

func (n *Network) Start() {
	go n.startRPC()
	go n.dialPeers()
	go n.processDeferred()
	go n.processInbound()
	n.processOutbound()
}

// Broadcast sends the msg to all nodes in the network asynchronously.
// NOTE: this function returns immediately without waiting for the other nodes to respond
func (n *Network) Broadcast(msg *types.Message) {
	n.ingestOutbound([]*types.Message{msg})
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
func (n *Network) processInbound() {
	for {
		n.mu.Lock()
		for len(n.outbound) == 0 {
			n.inboundCond.Wait()
		}
		for id, msg := range n.inbound {
			retry := n.handleMsg(msg)
			if retry {
				n.deferred[id] = msg
			}
		}
		n.mu.Unlock()
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

func (n *Network) processOutbound() {
	for {
		n.mu.Lock()
		for len(n.outbound) == 0 {
			n.outboundCond.Wait()
		}
		msgs := make([]*types.Message, 0, len(n.outbound))
		for _, msg := range n.outbound {
			msgs = append(msgs, msg)
		}
		n.mu.Unlock()
		n.doBroadcast(msgs)
	}
}

func (n *Network) ingestInbound(msgs []*types.Message) {
	n.mu.Lock()
	for _, msg := range msgs {
		n.inbound[string(msg.Hash())] = msg
	}
	n.mu.Unlock()
	n.inboundCond.Broadcast()
}

func (n *Network) ingestOutbound(msgs []*types.Message) {
	n.mu.Lock()
	for _, msg := range msgs {
		n.outbound[string(msg.Hash())] = msg
	}
	n.mu.Unlock()
	n.outboundCond.Broadcast()
}

func (n *Network) doBroadcast(msgs []*types.Message) {
	_msgs := &types.Messages{Msgs: msgs}
	envelope := &types.Envelope{
		Msgs:      _msgs,
		NodeIndex: n.idx,
		Sig:       n.sign(_msgs),
	}

	var wg sync.WaitGroup
	for i, client := range n.clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.Send(context.Background(), envelope)
			if err != nil {
				log.Errorf("failed to send to peer %d: %s", i, err.Error())
			}
		}()
	}
	wg.Wait()

	return
}

func (n *Network) sign(msg proto.Message) []byte {
	bs, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return crypto.Sign(n.key, bs)
}
