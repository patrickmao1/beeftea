package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	log "github.com/sirupsen/logrus"
)

func main() {
	for i := 0; i < 5; i++ {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		log.Infof("priv %x", key.D.Bytes())
	}
}
