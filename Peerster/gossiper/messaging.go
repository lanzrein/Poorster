package gossiper

import (
	"go.dedis.ch/onet/log"
	"github.com/JohanLanzrein/Peerster/clusters"
	"strings"
)

//SendAnonymousMessage - handles anonymous message sending
func (g *Gossiper) SendAnonymousMessage(gp GossipPacket, anonLevel float64) error {

	errChan := make(chan error)

	if !g.IsInCluster{
		log.Error("Cannot send anonymous message - current node does not belong to any cluster.")
		return <- errChan
	}

	var receiver string
	// sending an anonymous private message
	if gp.Private != nil {
		receiver = gp.Private.Destination

		if !isNodeInCluster(g.Cluster, receiver) {
				log.Error("Cannot send anonymous message, node ", receiver, " is not in the cluster.")
				return <- errChan
		}

		if _, ok := g.Cluster.PublicKeys[receiver]; !ok {
				log.Error("Cannot send anonymous message, public key of node ", receiver, " is not available.")
				return <- errChan
		}

		// anonymize the origin
		gp.Private.Origin = ""
	}

	encryptedBytes := g.EncryptPacket(gp, receiver)
	anonMsg := AnonymousMessage{
		EncryptedContent: encryptedBytes,
		Receiver:         receiver,
		AnonymityLevel:   anonLevel,
		RouteToReceiver:  false,
	}
	packetToSend := GossipPacket{AnonymousMsg: &anonMsg}

	go g.ReceiveAnonymousMessage(packetToSend, errChan)
	return <-errChan
}


func isNodeInCluster(c clusters.Cluster, node string) bool {

	for _, m := range c.Members {
		if strings.Compare(m, node) == 0 {
			return true;
		}
	}

	return false;
}
