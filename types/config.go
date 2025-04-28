package types

import (
	"crypto/ecdsa"
	"github.com/patrickmao1/beeftea/utils"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Config struct {
	InitTime          time.Time
	RoundDuration     time.Duration
	ProposalDuration  time.Duration
	ProposalThreshold uint32

	Peers []*Peer

	// cache fields
	myIndex *uint32
}

func (c *Config) MyIndex() uint32 {
	if c.myIndex != nil {
		return *c.myIndex
	}
	myIP, err := utils.GetPrivateIP()
	if err != nil {
		log.Panic(err)
	}
	for i, peer := range c.Peers {
		ip := strings.Split(peer.URL, ":")[0]
		if ip == myIP.String() {
			idx := uint32(i)
			c.myIndex = &idx
			return idx
		}
	}
	log.Fatalf("can't find myself in peer list: my ip %s", myIP.String())
	return 0
}

func (c *Config) MyKey() *ecdsa.PrivateKey {
	return c.Peers[c.MyIndex()].Key
}

type Peer struct {
	URL string
	Key *ecdsa.PrivateKey
}
