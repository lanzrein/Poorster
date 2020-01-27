package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/JohanLanzrein/Peerster/gossiper"
	"github.com/stretchr/testify/assert"
	"go.dedis.ch/onet/log"
)

func TestAnonymousMessaging(t *testing.T) {
	// log.SetDebugVisible(2)
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

	g3, err := gossiper.NewGossiper("C", "8083", "127.0.0.1:5002", "127.0.0.1:5003", false, 10, "8002", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g4, err := gossiper.NewGossiper("C", "8084", "127.0.0.1:5003", "127.0.0.1:5004", false, 10, "8003", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g5, err := gossiper.NewGossiper("C", "8085", "127.0.0.1:5004", "127.0.0.1:5005", false, 10, "8004", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g6, err := gossiper.NewGossiper("C", "8086", "127.0.0.1:5005", "127.0.0.1:5000", false, 10, "8005", 10, 3, 5, false, 10, false)
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
	go func() {
		g4.Run()
	}()
	go func() {
		g5.Run()
	}()
	go func() {
		g6.Run()
	}()

	for g1.FindPath(g3.Name) == "" {
		log.Warn("Path not established yet. Waiting 5 more seconds")
		<-time.After((5 * time.Second))
	}

	for g1.FindPath(g2.Name) == "" {
		log.Warn("Path not established yet. Waiting 5 more seconds")
		<-time.After((5 * time.Second))
	}

	for g1.FindPath(g4.Name) == "" {
		log.Warn("Path not established yet. Waiting 5 more seconds")
		<-time.After((5 * time.Second))
	}

	for g1.FindPath(g5.Name) == "" {
		log.Warn("Path not established yet. Waiting 5 more seconds")
		<-time.After((5 * time.Second))
	}

	for g1.FindPath(g6.Name) == "" {
		log.Warn("Path not established yet. Waiting 5 more seconds")
		<-time.After((5 * time.Second))
	}

	//G1 joins a cluster...
	g1.InitCluster()
	//G2 asks to join it..
	<-time.After(1 * time.Second)
	g2.RequestJoining(g1.Name, *g1.Cluster.ClusterID)
	g3.RequestJoining(g1.Name, *g1.Cluster.ClusterID)
	g4.RequestJoining(g1.Name, *g1.Cluster.ClusterID)
	g5.RequestJoining(g1.Name, *g1.Cluster.ClusterID)
	g6.RequestJoining(g1.Name, *g1.Cluster.ClusterID)
	<-time.After(30 * time.Second)
	c6 := g6.Cluster
	c5 := g5.Cluster
	c4 := g4.Cluster
	c3 := g3.Cluster
	c2 := g2.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.ClusterID, c3.ClusterID)
	assert.Equal(t, c1.ClusterID, c4.ClusterID)
	assert.Equal(t, c1.ClusterID, c5.ClusterID)
	assert.Equal(t, c1.ClusterID, c6.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.MasterKey, c3.MasterKey)
	assert.Equal(t, c1.MasterKey, c4.MasterKey)
	assert.Equal(t, c1.MasterKey, c5.MasterKey)
	assert.Equal(t, c1.MasterKey, c6.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.Members, c3.Members)
	assert.Equal(t, c1.Members, c4.Members)
	assert.Equal(t, c1.Members, c5.Members)
	assert.Equal(t, c1.Members, c6.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)
	assert.Equal(t, c1.PublicKeys, c3.PublicKeys)
	assert.Equal(t, c1.PublicKeys, c4.PublicKeys)
	assert.Equal(t, c1.PublicKeys, c5.PublicKeys)
	assert.Equal(t, c1.PublicKeys, c6.PublicKeys)

	anonText := "Anonymous Message"
	g1.ClientSendAnonymousMessage(g3.Name, anonText, 0.5, false)
	_, _ = g3.ReplyToClient()

	<-time.After(3 * time.Second)
	b1, err := g3.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}

	exp := fmt.Sprint("ANONYMOUS contents ", anonText, "\n")
	assert.Equal(t, string(b1), exp)
}
