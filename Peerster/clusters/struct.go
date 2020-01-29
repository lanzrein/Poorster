package clusters

import (
	"github.com/JohanLanzrein/Peerster/ies"
	"github.com/Peerster_solo/gossiper"
	"go.dedis.ch/onet/log"
	"math/rand"

)

type Cluster struct {
	ClusterID  *uint64
	Members    []string        //know the members by name
	HeartBeats map[string]bool //Members for whom we have a heartbeat.
	MasterKey  ies.PublicKey
	PublicKeys map[string]ies.PublicKey //public keys of the member of the cluster
	Seed       uint64
	Counter    uint64 //Count how many times the seed was sampled.
	source     *rand.Rand
	Authorities []string
}


func (c *Cluster)IsAnAuthority(name string) bool {
	return gossiper.Contains(c.Authorities, name)
}

func (c *Cluster) AmountAuthorities() int {
	return len(c.Authorities)
}
func NewCluster(id *uint64, members []string, masterkey ies.PublicKey, publickey map[string]ies.PublicKey, seed uint64 ) Cluster {
	source := rand.New(rand.NewSource(int64(seed)))

	return Cluster{
		ClusterID:  id,
		Members:    members,
		MasterKey:  masterkey,
		PublicKeys: publickey,
		HeartBeats: make(map[string]bool),
		Seed : seed ,
		source : source,
		Counter: 0 ,
		Authorities : []string{},
	}
}


func InitCounter(c *Cluster){
	source := rand.New(rand.NewSource(int64(c.Seed)))
	//Sample it enough time to be "on same clock cycle" as the rest.
	for i := uint64(0) ; i <  c.Counter; i ++ {
		source.Intn(len(c.Members))
	}

	c.source = source
	return
}

func (c *Cluster)Clock() int {
	c.Counter ++
	log.Lvl1("Clocking..", c.Counter)
	return c.source.Intn(len(c.Members))
}