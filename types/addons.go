package types

import (
	"github.com/patrickmao1/beeftea/crypto"
	"github.com/patrickmao1/beeftea/utils"
	"math"
)

func (e *Envelope) Hash() []byte {
	return utils.MustHash(e)
}

func (p *Proposal) Score() uint32 {
	if p == nil {
		return math.MaxUint32
	}
	return crypto.RngFromProof(p.ProposerProof)
}

func (p *Proposal) Hash() []byte {
	return utils.MustHash(p)
}
