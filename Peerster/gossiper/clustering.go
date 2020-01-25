package gossiper

import (
	"github.com/JohanLanzrein/Peerster/clusters"
	"github.com/JohanLanzrein/Peerster/ies"
	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"
	"time"
)

const DEFAULTROLLOUT = 300
const DEFAULTHEARTBEAT = 5

//InitCluster the current gossiper creates a cluster where he is the sole member
func (g *Gossiper) InitCluster() {
	id := GenerateId()
	members := []string{g.Name}
	publickey := make(map[string]ies.PublicKey)
	publickey[g.Name] = g.Keypair.PublicKey
	masterkey := g.MasterKeyGen()
	cluster := clusters.NewCluster(id, members, masterkey, publickey)

	g.Cluster = cluster
	g.IsInCluster = true
	g.PrintInitCluster()
	go g.HeartbeatLoop()
}

func (g *Gossiper) RequestJoining(other string, clusterID uint64) {
	//send a request packet to the other gossiper

	addr := g.FindPath(other)
	publickey := g.Keypair.PublicKey
	req := RequestMessage{
		Origin:    g.Name,
		Recipient: other,
		PublicKey: publickey,
	}

	gp := GossipPacket{JoinRequest: &req}
	go g.SendTo(addr, gp)
	//then the "voting" system starts

}

func (g *Gossiper) HeartbeatLoop() {
	for {

		select {
		case <-time.After(time.Duration(g.HearbeatTimer) * time.Second):
			log.Lvl4(g.Name , "sending heartbeat")
			g.Cluster.HeartBeats[g.Name] = true
			go g.SendBroadcast("", false )

		case <-g.LeaveChan:
			log.Lvl1("Leaving cluster")
			return
		case <-time.After(time.Duration(g.RolloutTimer) * time.Second):
			log.Lvl4("Time for a rolllllllllout")
			g.KeyRollout(g.Cluster.Members[0]) //TODO for now its only the first member but in the future chose randomly

		}
	}
}

func (g *Gossiper) LeaveCluster() {
	//Stop the heartbeat loop
	g.PrintLeaveCluster()
	g.LeaveChan <- true
	g.RequestLeave()

	g.Cluster = clusters.Cluster{}
	//Send a message saying we want to leave.
	log.Lvl2("Sending leave message..TODO")
	return
}

func (g *Gossiper) SendBroadcast(text string, leave bool ) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   text,
	}
	data, err := protobuf.Encode(&rumor)
	if err != nil {
		log.Error("Could not encode the packet.. ", err)
		return
	}

	enc := ies.Encrypt(g.Cluster.MasterKey, data)
	bm := BroadcastMessage{
		ClusterID:   *g.Cluster.ClusterID,
		HopLimit:    g.HopLimit,
		Destination: "",
		Data:        enc,
		LeaveRequest:leave,
	}
	gp := GossipPacket{Broadcast: &bm}

	//Send to all member of the cluster.
	//This does not need to be anonymized as an attacker can in any case know who is in a cluster by joining it..
	for _, m := range g.Cluster.Members {
		if m == g.Name {
			continue
		}
		bm.Destination = m
		addr := g.FindPath(m)
		if addr == "" {
			continue
		}
		err := g.SendTo(addr, gp)
		if err != nil {
			log.Error("Error while sending to ", m, " : ", err)
		}
	}
}

func (g *Gossiper) ReceiveBroadcast(message BroadcastMessage) {
	if message.Destination != g.Name {
		//Send it further.
		message.HopLimit--
		log.Lvl2(g.Name, "forwarding to ", message.Destination)
		if message.HopLimit > 0 {
			addr := g.FindPath(message.Destination)
			if addr == "" {
				log.Error(g.Name, "Could not find path to ", message.Destination)
				return
			}
			gp := GossipPacket{Broadcast: &message}
			g.SendToRandom(gp)
		}
		return
	}
	if g.Cluster.ClusterID != nil && message.ClusterID == *g.Cluster.ClusterID {
		log.Lvl2("Got broadcast for my cluster")

		if message.Rollout {
			//Update for a rollout.
			log.Lvl2(g.Name, " received message for rollout")
			cluster := clusters.Cluster{}
			data := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			err := protobuf.Decode(data, &cluster)
			if err != nil {
				log.Error("Could not decode rollout info ", err )
			}

			g.UpdateFromRollout(cluster)

		} else if message.LeaveRequest{
			log.Lvl1("Got leave request")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}

			g.Cluster.HeartBeats[rumor.Origin] = false
			delete(g.Cluster.PublicKeys,rumor.Origin)
		} else {
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			if rumor.Text != "" && rumor.Origin != g.Name {
				//print the message
				g.PrintBroadcast(rumor)

			}
			//in any case add it to the map..
			g.Cluster.HeartBeats[rumor.Origin] = true
		}

	}

}

func (g *Gossiper) ReceiveJoinRequest(message RequestMessage) {
	if message.Recipient != g.Name {
		addr := g.FindPath(message.Recipient)
		if addr == "" {
			log.Error("Could not find a path to ", message.Recipient)
			return
		}
		gp := GossipPacket{JoinRequest: &message}
		err := g.SendTo(addr, gp)
		if err != nil {
			log.Error("Error : ", err)
			return
		}
		return
	}



	_, ok := g.Cluster.HeartBeats[message.Origin]
	if ok {
		//its an update message.
		log.Lvl2(g.Name , " got an update message")
		g.Cluster.PublicKeys[message.Origin] = message.PublicKey
		return

	}

	//TODO start e-voting protocol to decide if accept...

	//Once the decision has been taken we have the result..
	//For now we always accept.

	g.UpdateCluster(message)
	data, err := protobuf.Encode(&g.Cluster)
	//encrypt it ...

	ek := g.Keypair.KeyDerivation(&message.PublicKey)
	enc := ies.Encrypt(ek, data)

	reply := RequestReply{
		Accepted:           true,
		Recipient:          message.Origin,
		ClusterID:          *g.Cluster.ClusterID,
		EphemeralKey:       g.Keypair.PublicKey,
		ClusterInformation: enc,
	}

	//Update the cluster with the new member info.
	gp := GossipPacket{RequestReply: &reply}
	addr := g.FindPath(message.Origin)
	if addr == "" {
		log.Error("Could not find the address")
	}
	err = g.SendTo(addr, gp)
	if err != nil {
		log.Error("Error while sending reply to ", message.Origin, " : ", err)
	}

	return

}

func (g *Gossiper) ReceiveRequestReply(message RequestReply) {

	if message.Recipient != g.Name {
		addr := g.FindPath(message.Recipient)
		if addr == "" {
			log.Error("Could not find path to ", message.Recipient)
			return
		}
		gp := GossipPacket{RequestReply: &message}
		err := g.SendTo(addr, gp)
		if err != nil {
			log.Error("Could not send packet ", err)
		}
		return
	}

	//reply <- g.ReplyChan
	if !message.Accepted {
		g.PrintDeniedJoining(message.ClusterID)
		return
	}

	var cluster clusters.Cluster
	pk := ies.PublicKey(message.EphemeralKey)
	data := ies.Decrypt(g.Keypair.KeyDerivation(&pk), message.ClusterInformation)
	err := protobuf.Decode(data, &cluster)
	if err != nil {
		log.Error("Could not decode cluster information :", err)
	}
	g.PrintAcceptJoiningID(cluster)

	g.Cluster = cluster
	//Start the heartbeatloop immediately
	go g.HeartbeatLoop()
}

//Initiate the key rollout. Assume that the leader has been elected and he calls this method.
func (g *Gossiper) KeyRollout(leader string) {
	//Send a new key pair to the "leader"

	var err error
	g.Keypair, err = ies.GenerateKeyPair()
	g.Cluster.PublicKeys = make(map[string]ies.PublicKey)


	if err != nil {
		log.Error("Could not generate new keypair : ", err)
	}

	go func() {

		if leader != g.Name {
			//Request to join
			<-time.After(time.Second)
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
				log.Lvl1("Removing : ", m, " from cluster")
				delete(g.Cluster.PublicKeys, m)
			} else {
				log.Lvl1("Staying in cluster ", m)
				nextMembers = append(nextMembers, m)
			}



		}
		g.Cluster.Members = nextMembers
		log.Lvl2("New members for this key rollout : ", nextMembers)

		//Check if received all the keys from them
		for {
			<-time.After(time.Second)
			if len(g.Cluster.PublicKeys) == len(nextMembers) {
				//we got all the maps we can generate the master key and return
				log.Lvl2("Got all the members needed")
				err := g.AnnounceNewMasterKey()
				if err != nil{
					log.Error("Could not announce master key : ", err )
				}
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

	return kp.PublicKey
}

func (g *Gossiper) AnnounceNewMasterKey() error {
	old := g.Cluster.MasterKey
	g.Cluster.MasterKey = g.MasterKeyGen()

	cluster := g.Cluster
	data, err := protobuf.Encode(&cluster)
	if err != nil {
		log.Error("Could not encode cluster :", err)
		return err
	}
	cipher := ies.Encrypt(old, data)

	for _, member := range cluster.Members {
		//send them the new master key using the previous master key
		if member == g.Name {
			continue
		}
		addr := g.FindPath(member)

		bc := BroadcastMessage{
			ClusterID: *cluster.ClusterID,
			Destination:member ,
			HopLimit:  10,
			Rollout:   true,
			Data:      cipher,
		}
		gp := GossipPacket{Broadcast: &bc}
		go g.SendTo(addr, gp)
	}

	return nil
}

func (g *Gossiper) UpdateCluster(message RequestMessage) {
	g.Cluster.Members = append(g.Cluster.Members, message.Origin)
	g.Cluster.PublicKeys[message.Origin] = message.PublicKey
	g.Cluster.HeartBeats[message.Origin] = true
}

func (g *Gossiper) UpdateFromRollout(cluster clusters.Cluster) {
	log.Lvl3("Update information form a new cluster :O ")

	g.Cluster = cluster
	g.Cluster.HeartBeats = make(map[string]bool)

}

func (g *Gossiper) RequestLeave() {
	g.SendBroadcast("", true)
}
