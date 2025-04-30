package main

import (
	"math"
	"time"

	"github.com/patrickmao1/beeftea/consensus"
	"github.com/patrickmao1/beeftea/crypto"
	"github.com/patrickmao1/beeftea/types"
)

func main() {
	config := &types.Config{
		InitTime:          time.Unix(1745816400, 0), // 2025-4-28 00:00:00
		RoundDuration:     4 * time.Second,
		ProposalDuration:  1 * time.Second,
		ProposalThreshold: computeThreshold(5),
		Peers: []*types.Peer{
			{URL: "172.16.0.1:9090", Key: crypto.UnmarshalHex("c71e183d51e9fae1d4fc410ca16a17a3a89da8e105b0e108576e2a77133f87b0")},
			{URL: "172.16.0.2:9090", Key: crypto.UnmarshalHex("26c65dc72d016ebe50a5751c258d8ff3ddc3da40b5dcf7ac638619e041119b71")},
			{URL: "172.16.0.3:9090", Key: crypto.UnmarshalHex("6a7e2b2ee79a8444d489c900f0c32bd944c88530882c9348b0d477b825773956")},
			{URL: "172.16.0.4:9090", Key: crypto.UnmarshalHex("64d691d9af74ff28b23f38e49bedbae5aa5298933c477d60173a3990eb263481")},
			{URL: "172.16.0.5:9090", Key: crypto.UnmarshalHex("376bb541ff3c913ea6b07cc0c405b991d354e140b4c3b8908884b84ef1475984")},
		},
	}

	s := consensus.NewService(config)
	s.Start()
}

func computeThreshold(n int) uint32 {
	// T = f(N) such that the probability of no one proposing a block is 0.01
	// f(N) = 1 - e^(-4.60517/N),
	const constant = 4.60517
	t := 1.0 - math.Exp(-constant/float64(n))
	return uint32(float64(math.MaxUint32) * t)
}
