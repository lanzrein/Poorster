package test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/JohanLanzrein/Peerster/gossiper"
	"github.com/stretchr/testify/assert"
	"go.dedis.ch/onet/log"
)

func TestAnonymousMessages(t *testing.T) {
	// log.SetDebugVisible(1)
	//Test the anonymous messaging within a cluster
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 3, 5, true, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5002", false, 10, "8001", 10, 3, 5, true, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g3, err := gossiper.NewGossiper("C", "8083", "127.0.0.1:5002", "127.0.0.1:5000", false, 10, "8002", 10, 3, 5, true, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g4, err := gossiper.NewGossiper("D", "8084", "127.0.0.1:5003", "127.0.0.1:5000", false, 10, "8003", 10, 3, 5, true, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g5, err := gossiper.NewGossiper("E", "8085", "127.0.0.1:5004", "127.0.0.1:5005", false, 10, "8004", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g6, err := gossiper.NewGossiper("F", "8086", "127.0.0.1:5005", "127.0.0.1:5000", false, 10, "8005", 10, 3, 5, false, 10, false)
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
	g2.RequestJoining(g1.Name)
	g3.RequestJoining(g1.Name)
	<-time.After(1 * time.Second)
	g4.RequestJoining(g1.Name)
	g5.RequestJoining(g1.Name)
	<-time.After(1 * time.Second)
	g6.RequestJoining(g1.Name)
	// g5.RequestJoining(g1.Name, *g1.Cluster.ClusterID)
	// g6.RequestJoining(g1.Name, *g1.Cluster.ClusterID)
	<-time.After(15 * time.Second)
	sort.Strings(g1.Cluster.Members)
	sort.Strings(g2.Cluster.Members)
	sort.Strings(g3.Cluster.Members)
	sort.Strings(g4.Cluster.Members)
	sort.Strings(g5.Cluster.Members)
	sort.Strings(g6.Cluster.Members)
	go func() { idx := 0; g1.KeyRollout(g1.Cluster.Members[idx]) }()
	go func() { idx := 0; g2.KeyRollout(g2.Cluster.Members[idx]) }()

	go func() { idx := 0; g3.KeyRollout(g3.Cluster.Members[idx]) }()
	go func() { idx := 0; g4.KeyRollout(g4.Cluster.Members[idx]) }()

	go func() { idx := 0; g5.KeyRollout(g5.Cluster.Members[idx]) }()
	go func() { idx := 0; g6.KeyRollout(g6.Cluster.Members[idx]) }()

	<-time.After(5 * time.Second)
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

	// Anonymous message from gossiper_1 to gossiper_6
	anonText_1_6 := "Anonymous Message to 6"
	_, _ = g6.ReplyToClient()
	g1.ClientSendAnonymousMessage(g6.Name, anonText_1_6, 0.9, false)

	<-time.After(1 * time.Second)
	b1, err := g6.ReplyToClient()
	if err != nil {
		log.Error(err, "Could not get the buffer")
	}

	exp := fmt.Sprint("ANONYMOUS contents ", anonText_1_6, "\n")
	assert.Equal(t, string(b1), exp)

	// Anonymous message from gossiper_1 to gossiper_6
	anonText_2_5 := "Anonymous Message from 2 to 5"
	_, _ = g5.ReplyToClient()
	g2.ClientSendAnonymousMessage(g5.Name, anonText_2_5, 0.7, false)

	<-time.After(1 * time.Second)
	b2, err := g5.ReplyToClient()
	if err != nil {
		log.Error(err, "Could not get the buffer")
	}

	exp_2 := fmt.Sprint("ANONYMOUS contents ", anonText_2_5, "\n")
	assert.Equal(t, string(b2), exp_2)

	// Anonymous message from gossiper_4 to gossiper_3
	anonText_4_3 := "Anonymous Message from 4 to 3"
	_, _ = g3.ReplyToClient()
	g4.ClientSendAnonymousMessage(g3.Name, anonText_4_3, 0.5, false)

	<-time.After(1 * time.Second)
	b3, err := g3.ReplyToClient()
	if err != nil {
		log.Error(err, "Could not get the buffer")
	}

	exp_3 := fmt.Sprint("ANONYMOUS contents ", anonText_4_3, "\n")
	assert.Equal(t, string(b3), exp_3)

}
