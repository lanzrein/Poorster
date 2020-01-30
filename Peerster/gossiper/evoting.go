//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent
//evoting handles the evoting protocol and the different steps

package gossiper

import (
	"github.com/JohanLanzrein/Peerster/ies"
	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"
)

//BroadcastJoin sends a broadcast concerning the node that will ask to join.
func (g *Gossiper) BroadcastJoin(nodeToJoin string) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   nodeToJoin,
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
		JoinRequest: true,
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
	bm.Destination = g.Name
	g.ReceiveBroadcast(bm)
}


func (g *Gossiper) BroadcastExpel(nodeToExpel string) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   nodeToExpel,
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
		ExpelRequest: true,
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
	bm.Destination = g.Name
	g.ReceiveBroadcast(bm)
}


//BroadcastAccept broadcast to all parties that the gossiper accepts this node
func (g *Gossiper) BroadcastAccept(nodeToAccept string) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   nodeToAccept,
	}
	data, err := protobuf.Encode(&rumor)
	if err != nil {
		log.Error("Could not encode the packet.. ", err)
		return
	}

	enc := ies.Encrypt(g.Cluster.MasterKey, data)
	bm := BroadcastMessage{
		ClusterID:         *g.Cluster.ClusterID,
		HopLimit:          g.HopLimit,
		Destination:       "",
		Data:              enc,
		AcceptProposition: true,
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
	bm.Destination = g.Name
	g.ReceiveBroadcast(bm)
}

//BroadcastDeny broadcast to all members that this gossiper denies the node
func (g *Gossiper) BroadcastDeny(nodeToDeny string) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   nodeToDeny,
	}
	data, err := protobuf.Encode(&rumor)
	if err != nil {
		log.Error("Could not encode the packet.. ", err)
		return
	}

	enc := ies.Encrypt(g.Cluster.MasterKey, data)
	bm := BroadcastMessage{
		ClusterID:       *g.Cluster.ClusterID,
		HopLimit:        g.HopLimit,
		Destination:     "",
		Data:            enc,
		DenyProposition: true,
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
	bm.Destination = g.Name
	g.ReceiveBroadcast(bm)
}

func (g *Gossiper) BroadcastCollected(caseID string) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   caseID,
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
		CaseCompare: true,
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
	bm.Destination = g.Name
	g.ReceiveBroadcast(bm)
}

//BroadcastResults broadcasts the result of the evoting.
func (g *Gossiper) BroadcastResults(results []string) {
	rumor := RumorMessage{
		Origin:  g.Name,
		ID:      0,
		Results: results,
	}
	data, err := protobuf.Encode(&rumor)
	if err != nil {
		log.Error("Could not encode the packet.. ", err)
		return
	}

	enc := ies.Encrypt(g.Cluster.MasterKey, data)
	bm := BroadcastMessage{
		ClusterID:         *g.Cluster.ClusterID,
		HopLimit:          g.HopLimit,
		Destination:       "",
		Data:              enc,
		ResultsValidation: true,
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

//BroadcastDecision broadcast the decision of an authority.
func (g *Gossiper) BroadcastDecision(answer string) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   answer,
	}
	data, err := protobuf.Encode(&rumor)
	if err != nil {
		log.Error("Could not encode the packet.. ", err)
		return
	}

	enc := ies.Encrypt(g.Cluster.MasterKey, data)
	bm := BroadcastMessage{
		ClusterID:     *g.Cluster.ClusterID,
		HopLimit:      g.HopLimit,
		Destination:   "",
		Data:          enc,
		FinalDecision: true,
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
	bm.Destination = g.Name
	g.ReceiveBroadcast(bm)
}

func (g *Gossiper) BroadcastCancel(caseToCancel string) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   caseToCancel,
	}
	data, err := protobuf.Encode(&rumor)
	if err != nil {
		log.Error("Could not encode the packet.. ", err)
		return
	}

	enc := ies.Encrypt(g.Cluster.MasterKey, data)
	bm := BroadcastMessage{
		ClusterID:     *g.Cluster.ClusterID,
		HopLimit:      g.HopLimit,
		Destination:   "",
		Data:          enc,
		CancelRequest: true,
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
	bm.Destination = g.Name
	g.ReceiveBroadcast(bm)
}

func (g *Gossiper) BroadcastReset(caseReset string) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   caseReset,
	}
	data, err := protobuf.Encode(&rumor)
	if err != nil {
		log.Error("Could not encode the packet.. ", err)
		return
	}

	enc := ies.Encrypt(g.Cluster.MasterKey, data)
	bm := BroadcastMessage{
		ClusterID:       *g.Cluster.ClusterID,
		HopLimit:        g.HopLimit,
		Destination:     "",
		Data:            enc,
		ResetIndication: true,
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
	bm.Destination = g.Name
	g.ReceiveBroadcast(bm)
}

func (g *Gossiper) BroadcastAck(requestToResend string) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   requestToResend,
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
		AckResend:   true,
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
