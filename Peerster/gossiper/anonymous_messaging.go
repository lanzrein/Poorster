package gossiper

import (
	"math/rand"
	"strings"
	"time"

	"github.com/JohanLanzrein/Peerster/clusters"
	"go.dedis.ch/onet/log"
)

//ClientSendAnonymousMessage - handles anonymous message sending
func (g *Gossiper) ClientSendAnonymousMessage(gp GossipPacket, relayRate float64, fullAnonimity bool) error {

	errChan := make(chan error)

	if !g.IsInCluster {
		log.Error("Cannot send anonymous message - current node does not belong to any cluster.")
		return <-errChan
	}

	var receiver string
	// sending an anonymous private message
	if gp.Private != nil {
		receiver = gp.Private.Destination

		if !isNodeInCluster(g.Cluster, receiver) {
			log.Error("Cannot send anonymous message, node ", receiver, " is not in the cluster.")
			return <-errChan
		}

		if _, ok := g.Cluster.PublicKeys[receiver]; !ok {
			log.Error("Cannot send anonymous message, public key of node ", receiver, " is not available.")
			return <-errChan
		}

		// anonymize the origin
		if fullAnonimity {
			gp.Private.Origin = ""
		}
	}

	encryptedBytes := g.EncryptPacket(gp, receiver)
	anonMsg := AnonymousMessage{
		EncryptedContent: encryptedBytes,
		Receiver:         receiver,
		AnonymityLevel:   relayRate,
		RouteToReceiver:  false,
	}

	go g.ReceiveAnonymousMessage(&anonMsg, errChan)
	return <-errChan
}

// ReceiveAnonymousMessage - handles receiving a gossip packet with an anonymous message
func (g *Gossiper) ReceiveAnonymousMessage(anon *AnonymousMessage, errChan chan error) {
	routeBecauseOfCoinFlip := false
	if strings.Compare(anon.Receiver, g.Name) == 0 {
		// anonymous message is for us, decrypt it
		decryptedPacket, err := g.DecryptBytes(anon.EncryptedContent)
		if err != nil {
			log.Error(err)
		}

		if decryptedPacket.Private != nil {
			// we received an anonymous private message
			g.PrintAnonymousPrivateMessage(*decryptedPacket.Private)
		}
	} else if !anon.RouteToReceiver {
		// anonymous message is not for us
		// if we are still relaying the message, flip a weighted coin to relay or send to destination
		seed := rand.NewSource(time.Now().UnixNano())
		seededRand := rand.New(seed)
		randFloat := seededRand.Float64()

		packet := GossipPacket{AnonymousMsg: anon}
		if randFloat <= anon.AnonymityLevel {
			// if the random float is less than the desired anonimity level,
			//		pick a random neighbor and relay the message to them
			addr := g.SendToRandom(packet)
			log.Lvl2("Relaying the message to : ", addr)
		} else {
			routeBecauseOfCoinFlip = true
		}

		// if after flip we decided to route to destination or if the packet is already
		//	being routed to the destination (e.g. a node before us flipped a coint to route it)
		if routeBecauseOfCoinFlip || anon.RouteToReceiver {
			addr := g.FindPath(anon.Receiver)
			if addr == "" {
				//we do not know this peer we stop here
				errChan <- nil
				return
			}
			log.Lvl2("Sending the anonymous message to it's destination: ", addr)
			errChan <- g.SendTo(addr, packet)
			return
		}
	}
}

func isNodeInCluster(c clusters.Cluster, node string) bool {

	for _, m := range c.Members {
		if strings.Compare(m, node) == 0 {
			return true
		}
	}

	return false
}
