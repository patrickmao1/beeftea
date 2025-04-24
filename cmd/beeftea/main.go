package main

import "github.com/patrickmao1/beeftea/consensus"

func main() {
	s := consensus.NewService()
	s.Start()
}
