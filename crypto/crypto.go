package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"golang.org/x/crypto/blake2b"
	"math/big"
)

func VRF(key *ecdsa.PrivateKey, seed []byte) (rng uint32, proof []byte) {
	sig := Sign(key, seed)
	return RngFromProof(sig), sig
}

func RngFromProof(proof []byte) uint32 {
	rng := blake2b.Sum256(proof)
	return uint32(new(big.Int).SetBytes(rng[:32]).Uint64())
}

func GenKey() *ecdsa.PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
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

func Marshal(key *ecdsa.PrivateKey) []byte {
	return key.D.Bytes()
}

func UnmarshalHex(hx string) (key *ecdsa.PrivateKey) {
	bs, err := hex.DecodeString(hx)
	if err != nil {
		panic(err)
	}
	return Unmarshal(bs)
}

func Unmarshal(bs []byte) (key *ecdsa.PrivateKey) {
	if len(bs) != 32 {
		panic("bad key length")
	}
	key = new(ecdsa.PrivateKey)
	key.Curve = elliptic.P256()
	key.D = new(big.Int).SetBytes(bs)
	key.PublicKey.X, key.PublicKey.Y = key.Curve.ScalarBaseMult(bs)
	return key
}
