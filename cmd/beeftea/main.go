package main

import (
	"fmt"
	"github.com/patrickmao1/beeftea/consensus"
	"github.com/patrickmao1/beeftea/crypto"
	"github.com/patrickmao1/beeftea/types"
	"math"
	"time"
)

func main() {
	config := &types.Config{
		InitTime:          time.Unix(1745816400, 0), // 2025-4-28 00:00:00
		RoundDuration:     4 * time.Second,
		ProposalDuration:  1 * time.Second,
		ProposalThreshold: computeThreshold(5),
	}
	for i := 1; i <= 5; i++ {
		config.Peers = append(config.Peers, &types.Peer{
			URL: fmt.Sprintf("172.16.0.%d:8080", i),
			Key: crypto.GenKeyDeterministic(int64(i)),
		})
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
