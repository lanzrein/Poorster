package test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/JohanLanzrein/Peerster/gossiper"
	"github.com/JohanLanzrein/Peerster/ies"
	"github.com/stretchr/testify/assert"
	"go.dedis.ch/onet/log"
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
	log.Lvl2("Cluster ", cluster.ClusterID, " members : ", cluster.Members, "master key :", cluster.MasterKey)

}

func TestJoinCluster(t *testing.T) {
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 2, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5001", false, 10, "8001", 10, 2, 5, false, 10, false)
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
	g2.RequestJoining(g1.Name)
	<-time.After(3 * time.Second)
	log.Lvl1("Checking if g2 joined the cluster.")

	c2 := g2.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)

}

func TestBroadcastCluster(t *testing.T) {
	//Test the broadcasting in the cluster
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5002", false, 10, "8001", 10, 3, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	g3, err := gossiper.NewGossiper("C", "8082", "127.0.0.1:5002", "127.0.0.1:5001", false, 10, "8002", 10, 3, 5, false, 10, false)
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
	//G1 joins a cluster...
	g1.InitCluster()
	//G2 asks to join it..
	<-time.After(1 * time.Second)
	g3.RequestJoining(g1.Name)
	<-time.After(5 * time.Second)
	c3 := g3.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c3.ClusterID)
	assert.Equal(t, c1.MasterKey, c3.MasterKey)
	assert.Equal(t, c1.Members, c3.Members)
	assert.Equal(t, c1.PublicKeys, c3.PublicKeys)
	text := "Hello from g3!"
	_, _ = g1.ReplyToClient()

	g3.SendBroadcast(text, false)

	<-time.After(3 * time.Second)
	b1, err := g1.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}

	exp := fmt.Sprint("Broadcast origin ", g3.Name, " contents ", text, "\n")
	assert.Equal(t, string(b1), exp)

}

func TestLeavingCluster(t *testing.T) {
	//Test if member leave the cluster properly.
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 2, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5001", false, 10, "8001", 10, 2, 5, false, 10, false)
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
		<-time.After(5 * time.Second)
	}
	//G1 joins a cluster...
	g1.InitCluster()
	//G2 asks to join it..
	g2.RequestJoining(g1.Name)
	<-time.After(3 * time.Second)
	log.Lvl1("Checking if g2 joined the cluster.")

	c2 := g2.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)

	g2.LeaveCluster()
	<-time.After(6 * time.Second)
	go func() { g1.KeyRollout(g1.Name) }()
	<-time.After(4 * time.Second)
	c1 = g1.Cluster
	assert.Len(t, c1.Members, 1)
	assert.Equal(t, c1.Members, []string{g1.Name})

	//Check if g2.Cluster is clear..
	c2 = g2.Cluster
	assert.Nil(t, c2.ClusterID)

}

func TestKeyRollout(t *testing.T) {
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 2, 5, false, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5001", false, 10, "8001", 10, 2, 5, false, 10, false)
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
		<-time.After(5 * time.Second)
	}
	//G1 joins a cluster...
	g1.InitCluster()
	//G2 asks to join it..
	g2.RequestJoining(g1.Name)
	<-time.After(3 * time.Second)
	log.Lvl1("Checking if g2 joined the cluster.")

	c2 := g2.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)
	var old ies.PublicKey
	copy(old, c1.MasterKey)
	g2.SendBroadcast("Hello world", false)

	<-time.After(6 * time.Second)

	go func() { g1.KeyRollout(g1.Name) }()
	go func() { g2.KeyRollout(g1.Name) }()

	<-time.After(3 * time.Second)
	c1 = g1.Cluster
	c2 = g2.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)
	assert.NotEqual(t, old, c1.MasterKey)

}
