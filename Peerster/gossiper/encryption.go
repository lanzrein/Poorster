package gossiper

import (
	"bytes"
	"errors"
	"github.com/JohanLanzrein/Peerster/ies"
	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"
)

func (g* Gossiper)GenerateKeys()(err error){

	g.Keypair , err = ies.GenerateKeyPair()
	return

}

func (g* Gossiper)EncryptPacket(packet GossipPacket, receiver string)[]byte{
	pubkey := g.Cluster.PublicKeys[receiver]
	//sample a new key
	kp, err := ies.GenerateKeyPair()
	if err != nil{
		log.Error("Error generating ephemeral key : " , err)
		return []byte{}
	}

	shared := kp.KeyDerivation(&pubkey)
	//todo should we store the shared key until a key rollout ? could be interesting...
	data, err := protobuf.Encode(&packet)
	if err != nil{
		log.Error("Could not encode packet : ", err)
		return []byte{}
	}
	var buf bytes.Buffer
	buf.Write(kp.PublicKey)
	log.Lvlf3("Public key bytes : %x", kp.PublicKey)

	cipher := ies.Encrypt(shared,data)
	buf.Write(cipher)
	return buf.Bytes()
}

//DecryptBytes tries to decrypt bytes and convert them to a GossipPacket
func (g *Gossiper) DecryptBytes(data []byte)(GossipPacket, error){
	if len(data) < 32{
		return GossipPacket{}, errors.New("unsufficient size byte array ")
	}

	ephKey  := ies.PublicKeyFromBytes(data[:32])
	log.Lvlf3("Retrieved public key bytes : %x",ephKey)
	cipher := data[32:]
	shared := g.Keypair.KeyDerivation(&ephKey)

	pt := ies.Decrypt(shared, cipher)
	var gp GossipPacket
	err := protobuf.Decode(pt, &gp)
	if err != nil{
		log.Error("Could not decode packet :" , err)
		return GossipPacket{}, err
	}

	return gp, nil

}