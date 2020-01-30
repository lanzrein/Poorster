//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent
//clusters utility functions for the cluster
package clusters

import (
	"math/rand"

	"github.com/JohanLanzrein/Peerster/ies"
	"go.dedis.ch/onet/log"
)

//Cluster A structure containing values for the cluster.
type Cluster struct {
	ClusterID   *uint64
	Members     []string        //know the members by name
	HeartBeats  map[string]bool //Members for whom we have a heartbeat.
	MasterKey   ies.PublicKey
	PublicKeys  map[string]ies.PublicKey //public keys of the member of the cluster
	Seed        uint64
	Counter     uint64 //Count how many times the seed was sampled.
	source      *rand.Rand
	Authorities []string
}

//IsAnAuthority returns true iff the name is c.Authorities
func (c *Cluster) IsAnAuthority(name string) bool {
	return contains(c.Authorities, name)
}

func contains(xs []string, val string) bool {
	for _, x := range xs {
		if val == x {
			return true
		}
	}
	return false
}

//AmountAuthorities returns the number of authorities
func (c *Cluster) AmountAuthorities() int {
	return len(c.Authorities)
}

//NewCluster returns a new cluster
func NewCluster(id *uint64, members []string, masterkey ies.PublicKey, publickey map[string]ies.PublicKey, seed uint64, authorities []string) Cluster {
	source := rand.New(rand.NewSource(int64(seed)))

	return Cluster{
		ClusterID:   id,
		Members:     members,
		MasterKey:   masterkey,
		PublicKeys:  publickey,
		HeartBeats:  make(map[string]bool),
		Seed:        seed,
		source:      source,
		Counter:     0,
		Authorities: authorities,
	}
}

//InitCounter initiliazes the source for the cluster to make sure it is synchronized
func InitCounter(c *Cluster) {
	source := rand.New(rand.NewSource(int64(c.Seed)))
	//Sample it enough time to be "on same clock cycle" as the rest.
	for i := uint64(0); i < c.Counter; i++ {
		source.Intn(len(c.Members))
	}

	c.source = source
	return
}

func (c *Cluster) Clock() int {
	c.Counter++
	log.Lvl2("Clocking..", c.Counter)
	return c.source.Intn(len(c.Members))
}
