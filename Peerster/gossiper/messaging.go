package gossiper

import "errors"

func (g *Gossiper)SendAnonymousMessage(name string)error{
	//Derive the key.
	publicKey, ok  := g.Cluster.PublicKeys[name]
	if !ok{
		return errors.New("non existing member")
	}

	derived := g.Keypair.KeyDerivation(&publicKey)



}