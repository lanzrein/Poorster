package gossiper

import (
	"github.com/JohanLanzrein/Peerster/clusters"
	"github.com/JohanLanzrein/Peerster/ies"
	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"
	"time"
)

//InitCluster the current gossiper creates a cluster where he is the sole member
func (g *Gossiper) InitCluster() {
	id := GenerateId()
	members := []string{g.Name}
	publickey := make(map[string]ies.PublicKey)
	publickey[g.Name] = g.Keypair.PublicKey
	masterkey := g.MasterKeyGen()
	cluster := clusters.NewCluster(id, members, masterkey, publickey)

	g.Cluster = cluster
}


func (g *Gossiper) RequestJoining(other string, clusterID uint64) {
	//send a request packet to the other gossiper
	addr := g.FindPath(other)
	publickey := g.Keypair.PublicKey
	req := RequestMessage{
		Origin:    g.Name,
		PublicKey: publickey,
	}

	gp := GossipPacket{JoinRequest: &req}
	go g.SendTo(addr, gp)
	//then the "voting" system starts


}


func (g *Gossiper) HeartbeatLoop() {
	for {

		select {
		case <-time.After(120 * time.Second):
			log.Lvl2("Sending heartbeat")
			//Todo broadcast to cluster
			go g.SendBroadcast("")

		case <-g.LeaveChan:
			log.Lvl2("Leaving cluster")
			return
		}
	}
}

func (g *Gossiper) LeaveCluster() {
	//Stop the heartbeat loop
	g.LeaveChan <- true
	g.Cluster = clusters.Cluster{}
	return
}

func (g *Gossiper) SendBroadcast(text string) {
	enc := ies.Encrypt(g.Cluster.MasterKey, []byte(text))
	bm := BroadcastMessage{
		ClusterID: *g.Cluster.ClusterID,
		Data:      enc,
	}
	gp := GossipPacket{Broadcast: &bm}

	//Send to all member of the cluster.
	//This does not need to be anonymized as an attacker can in any case know who is in a cluster by joining it..
	for _, m := range g.Cluster.Members {
		addr := g.FindPath(m)
		err := g.SendTo(addr, gp)
		if err != nil {
			log.Error("Error while sending to ", m, " : ", err)
		}
	}
}

func (g *Gossiper) ReceiveBroadcast(message BroadcastMessage) {
	decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
	rumor := RumorMessage{}
	err := protobuf.Decode(decrypted, &rumor)
	if err != nil {
		log.Error("Error decoding packet : ", err)
	}
	if rumor.Text != "" {
		//print the message
		g.PrintBroadcast(rumor)

	}
	//in any case add it to the map..
	g.Cluster.HeartBeats[rumor.Origin] = true

}
func (g *Gossiper) ReceiveJoinRequest(message RequestMessage) {
	_, ok := g.Cluster.PublicKeys[message.Origin]
	if ok {
		//its an update message.
		log.Lvl2("Update message")
		g.Cluster.PublicKeys[message.Origin] = message.PublicKey

	}

	//TODO start e-voting protocol to decide if accept...

	//Once the decision has been taken we have the result..
	//For now we always accept.
	reply := RequestReply{
		Accepted:           true,
		ClusterInformation: g.Cluster,
	}

	gp := GossipPacket{RequestReply: &reply}
	err := g.SendTo(g.FindPath(message.Origin), gp)
	if err != nil {
		log.Error("Error while sending reply to ", message.Origin, " : ", err)
	}

	return
}
func (g *Gossiper) ReceiveRequestReply(message RequestReply){
	var reply RequestReply
	//reply <- g.ReplyChan
	if !reply.Accepted {
		g.PrintDeniedJoining(*message.ClusterInformation.ClusterID)
		return
	}

	g.PrintAcceptJoiningID(reply.ClusterInformation)

	g.Cluster = reply.ClusterInformation
	//Start the heartbeatloop immediately
	go g.HeartbeatLoop()
}


//Initiate the key rollout. Assume that the leader has been elected and he calls this method.
func (g *Gossiper) KeyRollout(leader string) {
	//Send a new key pair to the "leader"
	g.Cluster.PublicKeys = make(map[string]ies.PublicKey)
	g.Cluster.HeartBeats = make(map[string]bool)

	go func() {
		var err error
		g.Keypair, err = ies.GenerateKeyPair()
		if err != nil {
			log.Error("Could not generate new keypair : ", err)
		}
		if leader == g.Name {
			//Request to join
			go g.RequestJoining(leader, *g.Cluster.ClusterID)
		}
		g.Cluster.PublicKeys[g.Name] = g.Keypair.PublicKey

	}()

	//leader does the rest.
	if leader == g.Name {

		//Check who is still in the cluster.
		var nextMembers []string
		for _, m := range g.Cluster.Members {
			flag, ok := g.Cluster.HeartBeats[m]
			if !ok || !flag {
				//he wants to be removed
				log.Lvl3("Removing : ", m, " from cluster")
				delete(g.Cluster.PublicKeys, m)
			} else {
				nextMembers = append(nextMembers, m)
			}
		}
		log.Lvl3("New members for this key rollout : ", nextMembers)

		//Check if received all the keys from them
		for {
			<-time.After(time.Second)
			if len(g.Cluster.PublicKeys) == len(nextMembers) {
				//we got all the maps we can generate the master key and return
				g.Cluster.MasterKey = g.MasterKeyGen()
				return
			}
		}

	}

}



func (g *Gossiper) MasterKeyGen() ies.PublicKey {
	kp, err := ies.GenerateKeyPair()
	if err != nil {
		log.Error("Could not generate key pair :", err)
	}

	//need to "announce" this is the new master key to the cluster.
	err = AnnounceNewMasterKey(g, &kp.PublicKey)
	if err != nil {
		log.Error("Could not announce new master key : ", err)
	}

	return kp.PublicKey
}
func AnnounceNewMasterKey(g *Gossiper, public *ies.PublicKey) error {
	cluster := g.Cluster
	data := []byte(*public)

	for _, member := range cluster.Members {
		//send them the new master key using the previous master key
		if member == g.Name {
			continue
		}
		addr := g.FindPath(member)
		cipher := ies.Encrypt(cluster.MasterKey, data)
		bc := BroadcastMessage{
			ClusterID: *cluster.ClusterID,
			Data:      cipher,
		}
		gp := GossipPacket{Broadcast: &bc}
		go g.SendTo(addr, gp)
	}

	return nil
}
