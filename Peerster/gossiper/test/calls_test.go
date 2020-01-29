package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/JohanLanzrein/Peerster/gossiper"
	"github.com/stretchr/testify/assert"
	"go.dedis.ch/onet/log"
)

func TestCallFunctionality(t *testing.T) {
	// log.SetDebugVisible(2)
	//Test the anonymous messaging within a cluster
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5000", false, 10, "8001", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		g1.Run()
	}()
	go func() {
		g2.Run()
	}()

	for g1.FindPath(g2.Name) == "" {
		log.Warn("Path not established yet. Waiting 5 more seconds")
		<-time.After((5 * time.Second))
	}

	//G1 joins a cluster...
	g1.InitCluster()
	//G2 asks to join it..
	<-time.After(1 * time.Second)
	g2.RequestJoining(g1.Name)
	<-time.After(20 * time.Second)
	c2 := g2.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)

	g1.ClientSendCallRequest(g2.Name)
	<-time.After(3 * time.Second)
	b1, err := g2.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}
	exp := fmt.Sprintf("CALL REQUEST from %s\n", g1.Name)
	assert.Equal(t, string(b1), exp)
}
