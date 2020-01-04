package clusters

import (
	"github.com/JohanLanzrein/Peerster/ies"
)

type Cluster struct {
	ClusterID  *uint64
	Members    []string        //know the members by name
	HeartBeats map[string]bool //Members for whom we have a heartbeat.
	MasterKey  ies.PublicKey
	PublicKeys map[string]ies.PublicKey //public keys of the member of the cluster

}

func NewCluster(id *uint64, members []string, masterkey ies.PublicKey, publickey map[string]ies.PublicKey) Cluster {
	return Cluster{
		ClusterID:  id,
		Members:    members,
		MasterKey:  masterkey,
		PublicKeys: publickey,
	}
}
