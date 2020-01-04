package clusters

import (
	"github.com/JohanLanzrein/Peerster/gossiper"
	"github.com/JohanLanzrein/Peerster/ies"
	"go.dedis.ch/protobuf"
)

type Cluster struct {
	ClusterID  *uint64
	Members    []string        //know the members by name
	HeartBeats map[string]bool //Members for whom we have a heartbeat.
	MasterKey  ies.PublicKey
	PublicKeys map[string]ies.PublicKey //public keys of the member of the cluster

}

func NewCluster(id *uint64, members []string, masterkey ies.PublicKey, publickey map[string]ies.PublicKey) clusters.Cluster {
	return Cluster{
		ClusterID:  id,
		Members:    members,
		MasterKey:  masterkey,
		PublicKeys: publickey,
	}
}

func (cluster Cluster) AnnounceNewMasterKey(g *gossiper.Gossiper, public *ies.PublicKey) error {
	data, err := protobuf.Encode(public)
	if err != nil {
		return err
	}

	for _, member := range cluster.Members {
		//send them the new master key using the previous master key
		if member == g.Name {
			continue
		}
		addr := g.FindPath(member)
		cipher := ies.Encrypt(cluster.MasterKey, data)
		bc := gossiper.BroadcastMessage{
			ClusterID: *cluster.ClusterID,
			Data:      cipher,
		}
		gp := gossiper.GossipPacket{Broadcast: &bc}
		go g.SendTo(addr, gp)
	}

	return nil
}
