package crypto

import (
	"crypto/ecdsa"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	key := GenKey()
	msg := []byte("hello world")
	sig := Sign(key, msg)
	pass := Verify(&key.PublicKey, msg, sig)
	require.True(t, pass)
}

func TestSerde(t *testing.T) {
	key := GenKey()
	msg := []byte("hello world")
	sig := Sign(key, msg)
	pass := Verify(&key.PublicKey, msg, sig)
	require.True(t, pass)

	bs := Marshal(key)
	key = &ecdsa.PrivateKey{}
	key = Unmarshal(bs)

	pass = Verify(&key.PublicKey, msg, sig)
	require.True(t, pass)
}
