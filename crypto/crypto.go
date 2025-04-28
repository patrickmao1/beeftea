package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"golang.org/x/crypto/blake2b"
	"math/big"
	mathrand "math/rand"
)

func VRF(key *ecdsa.PrivateKey, seed []byte) (rng uint32, proof []byte) {
	sig := Sign(key, seed)
	return RngFromProof(sig), sig
}

func RngFromProof(proof []byte) uint32 {
	rng := blake2b.Sum256(proof)
	return uint32(new(big.Int).SetBytes(rng[:32]).Uint64())
}

func GenKeyDeterministic(seed int64) *ecdsa.PrivateKey {
	rng := mathrand.New(mathrand.NewSource(seed))
	key, err := ecdsa.GenerateKey(elliptic.P256(), rng)
	if err != nil {
		panic(err)
	}
	return key
}

func Sign(key *ecdsa.PrivateKey, msg []byte) []byte {
	hash := blake2b.Sum256(msg)
	sig, err := ecdsa.SignASN1(rand.Reader, key, hash[:])
	if err != nil {
		panic(err)
	}
	return sig
}

func Verify(key *ecdsa.PublicKey, msg []byte, sig []byte) bool {
	hash := blake2b.Sum256(msg)
	return ecdsa.VerifyASN1(key, hash[:], sig)
}
