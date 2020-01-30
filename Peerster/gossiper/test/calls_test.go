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

func TestCalling(t *testing.T) {
	// log.SetDebugVisible(1)
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 3, 5, true, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5000", false, 10, "8001", 10, 3, 5, true, 10, false)
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
	<-time.After(1 * time.Second)
	<-time.After(15 * time.Second)
	sort.Strings(g1.Cluster.Members)
	sort.Strings(g2.Cluster.Members)
	go func() { idx := 0; g1.KeyRollout(g1.Cluster.Members[idx]) }()
	go func() { idx := 0; g2.KeyRollout(g2.Cluster.Members[idx]) }()

	<-time.After(5 * time.Second)
	c2 := g2.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)

	_, _ = g2.ReplyToClient()
	g1.ClientSendCallRequest(g2.Name)
	<-time.After(1 * time.Second)
	b1, err := g2.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}
	exp := fmt.Sprintf("CALL REQUEST from %s\n", g1.Name)
	assert.Equal(t, string(b1), exp)
}

// Testing pick up functionality
func TestPickUp(t *testing.T) {
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 3, 5, true, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5000", false, 10, "8001", 10, 3, 5, true, 10, false)
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
	<-time.After(1 * time.Second)
	<-time.After(15 * time.Second)
	sort.Strings(g1.Cluster.Members)
	sort.Strings(g2.Cluster.Members)
	go func() { idx := 0; g1.KeyRollout(g1.Cluster.Members[idx]) }()
	go func() { idx := 0; g2.KeyRollout(g2.Cluster.Members[idx]) }()

	<-time.After(5 * time.Second)
	c2 := g2.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)

	_, _ = g2.ReplyToClient()
	g1.ClientSendCallRequest(g2.Name)
	<-time.After(1 * time.Second)
	b1, err := g2.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}
	exp := fmt.Sprintf("CALL REQUEST from %s\n", g1.Name)
	assert.Equal(t, string(b1), exp)

	_, _ = g1.ReplyToClient()
	callResp := gossiper.CallResponse{Origin: g2.Name, Destination: g1.Name, Status: gossiper.Accept}
	g2.SendCallResponse(callResp)
	<-time.After(1 * time.Second)
	b2, err := g1.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}
	exp_2 := fmt.Sprintf("CALL ACCEPTED by %s\n", g2.Name)
	assert.Equal(t, string(b2), exp_2)
	assert.Equal(t, g2.CallStatus.InCall, true)
	assert.Equal(t, g2.CallStatus.ExpectingResponse, false)
	assert.Equal(t, g2.CallStatus.OtherParticipant, g2.Name)
}

// Testing pick up functionality
func TestDecline(t *testing.T) {
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 3, 5, true, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5000", false, 10, "8001", 10, 3, 5, true, 10, false)
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
	<-time.After(1 * time.Second)
	<-time.After(15 * time.Second)
	sort.Strings(g1.Cluster.Members)
	sort.Strings(g2.Cluster.Members)
	go func() { idx := 0; g1.KeyRollout(g1.Cluster.Members[idx]) }()
	go func() { idx := 0; g2.KeyRollout(g2.Cluster.Members[idx]) }()

	<-time.After(5 * time.Second)
	c2 := g2.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)

	_, _ = g2.ReplyToClient()
	g1.ClientSendCallRequest(g2.Name)
	<-time.After(1 * time.Second)
	b1, err := g2.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}
	exp := fmt.Sprintf("CALL REQUEST from %s\n", g1.Name)
	assert.Equal(t, string(b1), exp)

	_, _ = g1.ReplyToClient()
	callResp := gossiper.CallResponse{Origin: g2.Name, Destination: g1.Name, Status: gossiper.Decline}
	g2.SendCallResponse(callResp)
	<-time.After(1 * time.Second)
	b2, err := g1.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}
	exp_2 := fmt.Sprintf("CALL DECLINED by %s\n", g2.Name)
	assert.Equal(t, string(b2), exp_2)
	assert.Equal(t, g2.CallStatus.InCall, false)
	assert.Equal(t, g2.CallStatus.ExpectingResponse, false)
	assert.Equal(t, g2.CallStatus.OtherParticipant, "")
}

// Testing busy functionality
func TestBusy(t *testing.T) {
	g1, err := gossiper.NewGossiper("A", "8080", "127.0.0.1:5000", "127.0.0.1:5001", false, 10, "8000", 10, 3, 5, true, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	//suppose an other gossiper g2
	g2, err := gossiper.NewGossiper("B", "8081", "127.0.0.1:5001", "127.0.0.1:5000", false, 10, "8001", 10, 3, 5, true, 10, false)
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
	<-time.After(1 * time.Second)
	<-time.After(15 * time.Second)
	sort.Strings(g1.Cluster.Members)
	sort.Strings(g2.Cluster.Members)
	go func() { idx := 0; g1.KeyRollout(g1.Cluster.Members[idx]) }()
	go func() { idx := 0; g2.KeyRollout(g2.Cluster.Members[idx]) }()

	<-time.After(5 * time.Second)
	c2 := g2.Cluster
	c1 := g1.Cluster
	assert.Equal(t, c1.ClusterID, c2.ClusterID)
	assert.Equal(t, c1.MasterKey, c2.MasterKey)
	assert.Equal(t, c1.Members, c2.Members)
	assert.Equal(t, c1.PublicKeys, c2.PublicKeys)

	_, _ = g2.ReplyToClient()
	g1.ClientSendCallRequest(g2.Name)
	<-time.After(1 * time.Second)
	b1, err := g2.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}
	exp := fmt.Sprintf("CALL REQUEST from %s\n", g1.Name)
	assert.Equal(t, string(b1), exp)

	_, _ = g1.ReplyToClient()
	callResp := gossiper.CallResponse{Origin: g2.Name, Destination: g1.Name, Status: gossiper.Busy}
	g2.SendCallResponse(callResp)
	<-time.After(1 * time.Second)
	b2, err := g1.ReplyToClient()
	if err != nil {
		log.ErrFatal(err, "Could not get the buffer")
	}
	exp_2 := fmt.Sprintf("LINE BUSY by %s\n", g2.Name)
	assert.Equal(t, string(b2), exp_2)
}
