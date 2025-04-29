package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/patrickmao1/beeftea/crypto"
	"golang.org/x/crypto/blake2b"
	"math"
)

func (e *Envelope) Hash() []byte {
	if e == nil {
		panic("hash nil")
	}
	bs, err := proto.Marshal(e)
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
