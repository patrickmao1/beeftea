package utils

import (
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/blake2b"
)

func MustHash(msg proto.Message) []byte {
	if msg == nil {
		panic("cannot hash nil")
	}
	bs, err := proto.Marshal(msg)
	if err != nil {
		panic("failed to marshal proto message: " + err.Error())
	}
	hash := blake2b.Sum256(bs)
	return hash[:]
}
