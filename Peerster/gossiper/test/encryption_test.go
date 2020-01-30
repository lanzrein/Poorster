package test

import (
	"github.com/JohanLanzrein/Peerster/gossiper"
	"github.com/JohanLanzrein/Peerster/ies"
	"go.dedis.ch/onet/log"
	"testing"
	"time"
)

//func NewGossiper(Name string, UIPort string, gossipAddr string, gossipers string, simple bool, antiEntropy int, GUIPort string, rtimer int, N int, stubbornTimeout uint64, ackAll bool, hopLimit uint32, hw3ex3 bool) (Gossiper, error) {
func TestGossipPacketEncryption(t *testing.T) {
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "", false, 10, "8000", 10, 2, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	g1.Keypair, err = ies.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}

	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "", false, 10, "8001", 10, 2, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	g2.Keypair, err = ies.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}

	//adding the respective key pair manually..
	g2.Keys[g1.Name] = g1.Keypair.PublicKey
	g1.Keys[g2.Name] = g2.Keypair.PublicKey

	//g1 "sends" a packet to g2
	pm := gossiper.PrivateMessage{
		Origin:      g1.Name,
		ID:          0,
		Text:        "Hello from g1 ",
		Destination: g2.Name,
		HopLimit:    g1.HopLimit,
	}

	gp := gossiper.GossipPacket{Private: &pm}
	now := time.Now()
	cipher := g1.EncryptPacket(gp, g2.Name)
	log.Lvlf2("Cipher of len %d is : %x", len(cipher), cipher)
	//then c2 "receives" the cipher and tries to open it.

	gp1, err := g2.DecryptBytes(cipher)
	elapsed := time.Since(now)
	log.Lvl1("Time taken for encryption-decryption of a packet :", elapsed)
	if err != nil {
		t.Fatal(err)
	}

	g2.PrintPrivateMessage(*gp1.Private)

}
