package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/patrickmao1/beeftea/crypto"
	"golang.org/x/crypto/blake2b"
	"math"
)

func (m *Message) Hash() []byte {
	if m == nil {
		panic("nil message")
	}
	bs, err := proto.Marshal(m)
	if err != nil {
		panic("failed to marshal message: " + err.Error())
	}
	hash := blake2b.Sum256(bs)
	return hash[:]
}

func (p *Proposal) Score() uint32 {
	if p == nil {
		return math.MaxUint32
	}
	return crypto.RngFromProof(p.ProposerProof)
}
