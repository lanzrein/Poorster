package test

import (
	"testing"
	"time"

	"github.com/JohanLanzrein/Peerster/gossiper"
	"github.com/stretchr/testify/assert"
	"go.dedis.ch/onet/log"
)

func TestAnonymousMessaging(t *testing.T) {
	log.SetDebugVisible(4)
	//Test the anonymous messaging within a cluster
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5002", false, 10, "8001", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g3, err := gossiper.NewGossiper("C", "8082", "127.0.0.1:5002", "127.0.0.1:5000", false, 10, "8002", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		g1.Run()
	}()
	go func() {
		g2.Run()
	}()
	go func() {
		g3.Run()
	}()

	for g1.FindPath(g3.Name) == "" {
		log.Warn("Path not established yet. Waiting 5 more seconds")
		<-time.After((5 * time.Second))
	}

	for g1.FindPath(g2.Name) == "" {
		log.Warn("Path not established yet. Waiting 5 more seconds")
		<-time.After((5 * time.Second))
	}

	for g2.FindPath(g3.Name) == "" {
		log.Warn("Path not established yet. Waiting 5 more seconds")
		<-time.After((5 * time.Second))
	}

	//G1 joins a cluster...
	g1.InitCluster()
	//G2 asks to join it..
	<-time.After(1 * time.Second)
	g2.RequestJoining(g1.Name)
	g3.RequestJoining(g1.Name)
	<-time.After(15 * time.Second)
	//c3 := g3.Cluster
	//c2 := g2.Cluster
	//c1 := g1.Cluster
	//assert.Equal(t, c1.ClusterID, c2.ClusterID)
	//assert.Equal(t, c1.ClusterID, c3.ClusterID)
	//assert.Equal(t, c1.MasterKey, c2.MasterKey)
	//assert.Equal(t, c1.MasterKey, c3.MasterKey)
	//assert.Equal(t, c1.Members, c3.Members)
	//assert.Equal(t, c1.PublicKeys, c3.PublicKeys)
	//assert.Equal(t, c1.Members, c2.Members)
	//assert.Equal(t, c1.PublicKeys, c2.PublicKeys)

	//Check if counter etc are ok
	for i := 0; i < 30; i++ {
		assert.Equal(t, g1.Cluster.Counter, g2.Cluster.Counter)
		assert.Equal(t, g2.Cluster.Counter, g3.Cluster.Counter)
		cl1 := g1.Cluster.Clock()
		cl2 := g2.Cluster.Clock()
		cl3 := g3.Cluster.Clock()
		assert.Equal(t, cl1, cl2)
		assert.Equal(t, cl2, cl3)
	}

}
