package consensus

import (
	log "github.com/sirupsen/logrus"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// Start starts the main consensus loop and the RPC servers
func (s *Service) Start() {
	log.Infoln("starting consensus RPC")
	go s.startConsensusRPC()

	log.Infoln("starting external RPC")
	go s.startExternalRPC()

	log.Infoln("starting consensus main loop")

	//for {
	//}
}
