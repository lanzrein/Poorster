package test

import (
	"bytes"
	"github.com/JohanLanzrein/Peerster/gossiper"
	"github.com/stretchr/testify/assert"
	"go.dedis.ch/onet/log"
	"testing"
)

func TestInitializedCluster(t *testing.T) {
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "", false, 10, "8000", 10, 2, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g1.InitCluster()
	assert.Contains(t, g1.Cluster.Members, g1.Name)
	g1PK, ok := g1.Cluster.PublicKeys[g1.Name]
	if !ok {
		t.Fatal("PublicKeys of cluster does not contain the public key of g1")
	}

	if !bytes.Equal(g1PK, g1.Keypair.PublicKey) {
		t.Fatal("PublicKey in cluster does not equal public key of g1")
	}
	cluster := g1.Cluster
	log.Lvl1("Cluster ", cluster.ClusterID, " members : ", cluster.Members, "master key :", cluster.MasterKey)

}

func TestJoinCluster(t *testing.T) {
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "", false, 10, "8000", 10, 2, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "", false, 10, "8001", 10, 2, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//G1 joins a cluster...
	g1.InitCluster()
	//G2 asks to join it..
	g2.RequestJoining(g1.Name, *g1.Cluster.ClusterID)

}
