package types

type NodeConfig struct {
	IP         string
	Port       int
	PrivateKey []byte
	PublicKey  []byte
}
